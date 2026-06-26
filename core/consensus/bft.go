package consensus

import (
	"errors"
	"sync"

	"github.com/sureshKrishna05/aegisq-framework/core/aqx"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
)

type VoteType int

const (
	Prepare VoteType = iota
	Commit
)

type Vote struct {
	ValidatorID uint64
	BlockHash   [32]byte
	View        uint64
	Type        VoteType
	Signature   []byte // Added Signature for cryptographic proof
}

// computePayloadHash uses the AQX Encoder to deterministically pack the Vote
func (v *Vote) computePayloadHash() ([]byte, error) {
	e := aqx.AcquireEncoder()
	defer e.Release()

	// Deterministically encode according to AQX RFC rules
	e.UInt64(v.ValidatorID)
	e.FixedBytes(v.BlockHash[:])
	e.UInt64(v.View)
	e.UInt64(uint64(v.Type)) // Cast VoteType to uint64

	return crypto.Hash(e.Bytes()), nil
}

// Hash calculates the deterministic hash of the vote payload
func (v *Vote) Hash() ([]byte, error) {
	return v.computePayloadHash()
}

// Sign applies the post-quantum or classical signature to the binary payload
func (v *Vote) Sign(node *identity.NodeIdentity) error {
	hash, err := v.computePayloadHash()
	if err != nil {
		return err
	}

	sig, err := node.Sign(hash)
	if err != nil {
		return err
	}

	v.Signature = sig
	return nil
}

// Verify checks the signature against the strict binary payload
func (v *Vote) Verify(signer crypto.Signer, pubKey []byte) (bool, error) {
	hash, err := v.computePayloadHash()
	if err != nil {
		return false, err
	}

	return signer.Verify(pubKey, hash, v.Signature), nil
}

// SerializeAQX completely flattens the Vote (INCLUDING the signature)
func (v *Vote) SerializeAQX() []byte {
	e := aqx.AcquireEncoder()
	defer e.Release()

	e.UInt64(v.ValidatorID)
	e.FixedBytes(v.BlockHash[:])
	e.UInt64(v.View)
	e.UInt64(uint64(v.Type))
	e.BytesArray(v.Signature)

	out := make([]byte, len(e.Bytes()))
	copy(out, e.Bytes())
	return out
}

// DeserializeAQX acts as a constructor that reads raw AQX network bytes
func DeserializeAQX(data []byte) (*Vote, error) {
	d := aqx.NewDecoder(data)
	v := &Vote{}
	var err error

	if v.ValidatorID, err = d.UInt64(); err != nil {
		return nil, err
	}
	if blockHashBytes, err := d.FixedBytes(32); err != nil {
		return nil, err
	} else {
		copy(v.BlockHash[:], blockHashBytes)
	}

	if v.View, err = d.UInt64(); err != nil {
		return nil, err
	}

	var vType uint64
	if vType, err = d.UInt64(); err != nil {
		return nil, err
	}
	v.Type = VoteType(vType)

	if v.Signature, err = d.BytesArray(); err != nil {
		return nil, err
	}

	return v, nil
}

type VotePool struct {
	mu sync.Mutex

	// blockHash -> view -> voteType -> validatorID -> bool
	votes map[[32]byte]map[uint64]map[VoteType]map[uint64]bool

	// seenVotes: view -> voteType -> validatorID -> blockHash
	seenVotes map[uint64]map[VoteType]map[uint64][32]byte

	validatorSet *ValidatorSet
}

func NewVotePool(vs *ValidatorSet) *VotePool {
	return &VotePool{
		votes:        make(map[[32]byte]map[uint64]map[VoteType]map[uint64]bool),
		seenVotes:    make(map[uint64]map[VoteType]map[uint64][32]byte),
		validatorSet: vs,
	}
}

func (vp *VotePool) AddVote(v Vote) error {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	// 1️⃣ Authorization check
	if _, exists := vp.validatorSet.GetValidator(v.ValidatorID); !exists {
		return errors.New("unauthorized validator")
	}

	// 2️⃣ Initialize seenVotes structure
	if _, ok := vp.seenVotes[v.View]; !ok {
		vp.seenVotes[v.View] = make(map[VoteType]map[uint64][32]byte)
	}

	if _, ok := vp.seenVotes[v.View][v.Type]; !ok {
		vp.seenVotes[v.View][v.Type] = make(map[uint64][32]byte)
	}

	// 3️⃣ Equivocation prevention
	if existingHash, voted := vp.seenVotes[v.View][v.Type][v.ValidatorID]; voted {
		if existingHash == v.BlockHash {
			return errors.New("double vote detected")
		}
		return errors.New("equivocation detected")
	}

	// Record seen vote globally
	vp.seenVotes[v.View][v.Type][v.ValidatorID] = v.BlockHash

	// 4️⃣ Store vote per block
	if _, ok := vp.votes[v.BlockHash]; !ok {
		vp.votes[v.BlockHash] = make(map[uint64]map[VoteType]map[uint64]bool)
	}

	if _, ok := vp.votes[v.BlockHash][v.View]; !ok {
		vp.votes[v.BlockHash][v.View] = make(map[VoteType]map[uint64]bool)
	}

	if _, ok := vp.votes[v.BlockHash][v.View][v.Type]; !ok {
		vp.votes[v.BlockHash][v.View][v.Type] = make(map[uint64]bool)
	}

	vp.votes[v.BlockHash][v.View][v.Type][v.ValidatorID] = true

	return nil
}

func (vp *VotePool) HasQuorum(blockHash [32]byte, view uint64, voteType VoteType) bool {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	n := vp.validatorSet.Count()
	if n == 0 {
		return false
	}

	f := (n - 1) / 3
	quorum := 2*f + 1

	if _, ok := vp.votes[blockHash]; !ok {
		return false
	}

	if _, ok := vp.votes[blockHash][view]; !ok {
		return false
	}

	count := len(vp.votes[blockHash][view][voteType])
	return count >= quorum
}
