package consensus

import (
	"errors"
	"sync"
	"time"

	"github.com/sureshKrishna05/aegisq-framework/core/aqx"
)

type Finality struct {
	BlockHash      [32]byte
	View           uint64
	ValidatorCount uint64
	FinalizedAt    uint64
}

func (f *Finality) SerializeAQX() []byte {
	e := aqx.AcquireEncoder()
	defer e.Release()

	e.FixedBytes(f.BlockHash[:])
	e.UInt64(f.View)
	e.UInt64(f.ValidatorCount)
	e.UInt64(f.FinalizedAt)

	out := make([]byte, len(e.Bytes()))
	copy(out, e.Bytes())
	return out
}

func DeserializeFinalityAQX(data []byte) (*Finality, error) {
	d := aqx.NewDecoder(data)
	f := &Finality{}
	var err error

	if blockHashBytes, err := d.FixedBytes(32); err != nil {
		return nil, err
	} else {
		copy(f.BlockHash[:], blockHashBytes)
	}

	if f.View, err = d.UInt64(); err != nil {
		return nil, err
	}
	if f.ValidatorCount, err = d.UInt64(); err != nil {
		return nil, err
	}
	if f.FinalizedAt, err = d.UInt64(); err != nil {
		return nil, err
	}

	return f, nil
}

type FinalityEngine struct {
	mu sync.Mutex

	votePool *VotePool

	// height -> blockHash finalized
	finalized map[uint64][32]byte

	// blockHash -> prepared
	prepared map[[32]byte]bool

	// blockHash -> AQX Finality Proof
	proofs map[[32]byte]*Finality
}

func NewFinalityEngine(vp *VotePool) *FinalityEngine {
	return &FinalityEngine{
		votePool:  vp,
		finalized: make(map[uint64][32]byte),
		prepared:  make(map[[32]byte]bool),
		proofs:    make(map[[32]byte]*Finality),
	}
}

func (fe *FinalityEngine) TryPrepare(height uint64, hash [32]byte, view uint64) bool {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if fe.votePool.HasQuorum([32]byte{}, uint64(view), Prepare) {
		fe.prepared[hash] = true
		return true
	}

	return false
}

func (fe *FinalityEngine) TryCommit(height uint64, hash [32]byte, view uint64) error {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if !fe.prepared[hash] {
		return errors.New("block not prepared")
	}

	if !fe.votePool.HasQuorum([32]byte{}, uint64(view), Commit) {
		return errors.New("commit quorum not reached")
	}

	if _, exists := fe.finalized[height]; exists {
		return errors.New("height already finalized")
	}

	fe.finalized[height] = hash

	fe.proofs[hash] = &Finality{
		BlockHash:      hash,
		View: uint64(view),
		ValidatorCount: uint64(fe.votePool.validatorSet.Count()),
		FinalizedAt:    uint64(time.Now().Unix()),
	}

	return nil
}

func (fe *FinalityEngine) IsFinalized(height uint64, hash [32]byte) bool {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	finalHash, exists := fe.finalized[height]
	return exists && finalHash == hash
}

func (fe *FinalityEngine) GetProof(hash [32]byte) (*Finality, bool) {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	proof, exists := fe.proofs[hash]
	return proof, exists
}
