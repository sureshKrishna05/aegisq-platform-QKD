package block

import (
	"bytes"
	"errors"
	"time"

	"github.com/sureshKrishna05/aegisq-framework/core/aqx"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
	"golang.org/x/sync/errgroup"
)

type Block struct {
	Index        uint64
	View         uint64
	Timestamp    uint64
	PreviousHash []byte
	MerkleRoot   []byte
	Transactions []*transaction.Transaction
	TransactionHashes [][]byte // Added for Phase 3 Transaction Separation
	Hash         []byte
	ValidatorID  uint64
	Signature    []byte
}

func NewBlock(index uint64, view uint64, prevHash []byte, txs []*transaction.Transaction) *Block {
	return &Block{
		Index:        index,
		View: uint64(view),
		Timestamp:    uint64(time.Now().Unix()),
		PreviousHash: prevHash,
		Transactions: txs,
	}
}

// computeBlockHash uses the AQX Encoder to deterministically pack the block header
// WITHOUT the hash or signature, guaranteeing a perfectly stable hash.
func (b *Block) computeBlockHash() ([]byte, error) {

	if b.MerkleRoot == nil {
		return nil, errors.New("merkle root not set")
	}

	// 1. Grab a zero-allocation buffer from the AQX pool
	e := aqx.AcquireEncoder()
	defer e.Release()

	// 2. Deterministically encode the block header using explicit uint64
	e.UInt64(b.Index)
	e.UInt64(b.View)
	e.UInt64(b.Timestamp)
	e.FixedBytes(b.PreviousHash) // FixedBytes assumes 32-byte hashes (no length prefix)
	e.FixedBytes(b.MerkleRoot)

	// 3. Hash the canonical AQX bytes.
	return crypto.Hash(e.Bytes()), nil
}

func (b *Block) Finalize(node *identity.NodeIdentity) error {

	if len(b.Transactions) == 0 {
		return errors.New("block must contain transactions")
	}

	var txHashes [][]byte

	for _, tx := range b.Transactions {
		hash, err := tx.Hash()
		if err != nil {
			return err
		}
		txHashes = append(txHashes, hash)
	}

	b.MerkleRoot = ComputeMerkleRoot(txHashes)

	hash, err := b.computeBlockHash()
	if err != nil {
		return err
	}

	b.Hash = hash
	b.ValidatorID = node.ValidatorID

	signature, err := node.Sign(hash)
	if err != nil {
		return err
	}

	b.Signature = signature

	return nil
}

func (b *Block) Verify(signer crypto.Signer, publicKey []byte) (bool, error) {

	if b.Hash == nil {
		return false, errors.New("block hash missing")
	}

	// 1️⃣ Verify each transaction in PARALLEL
	// Performance Engineering: Distribute signature checks across CPU cores
	var g errgroup.Group
	for _, tx := range b.Transactions {
		tx := tx // Capture for goroutine
		g.Go(func() error {
			valid, err := tx.Verify(signer)
			if err != nil {
				return err
			}
			if !valid {
				return errors.New("invalid transaction signature")
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}

	// 2️⃣ Recompute Merkle root
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		hash, err := tx.Hash()
		if err != nil {
			return false, err
		}
		txHashes = append(txHashes, hash)
	}

	expectedMerkle := ComputeMerkleRoot(txHashes)

	if !bytes.Equal(expectedMerkle, b.MerkleRoot) {
		return false, nil
	}

	// 3️⃣ Recompute block header hash
	expectedHash, err := b.computeBlockHash()
	if err != nil {
		return false, err
	}

	if !bytes.Equal(expectedHash, b.Hash) {
		return false, nil
	}

	// 4️⃣ Verify block signature
	return signer.Verify(publicKey, b.Hash, b.Signature), nil
}

// =====================================================================
// AQX NETWORK & STORAGE LAYER
// =====================================================================

// SerializeAQX fully flattens the Block AND all its Transactions
// into a single, contiguous byte array for network transmission or BoltDB.
func (b *Block) SerializeAQX() []byte {
	e := aqx.AcquireEncoder()
	defer e.Release()

	// 1. Serialize the Header
	e.UInt64(b.Index)
	e.UInt64(b.View)
	e.UInt64(b.Timestamp)
	e.BytesArray(b.PreviousHash) // Use BytesArray here to include the length prefix for safe parsing
	e.BytesArray(b.MerkleRoot)
	e.BytesArray(b.Hash)
	e.UInt64(b.ValidatorID)
	e.BytesArray(b.Signature)

	// 2. Serialize the Transactions List as HASHES ONLY
	e.UInt32(uint32(len(b.Transactions)))
	for _, tx := range b.Transactions {
		hash, _ := tx.Hash()
		e.FixedBytes(hash)
	}

	out := make([]byte, len(e.Bytes()))
	copy(out, e.Bytes())
	return out
}

// DeserializeAQX reconstructs a full Block object and all nested Transactions
// from a raw AQX network payload.
func DeserializeAQX(data []byte) (*Block, error) {
	d := aqx.NewDecoder(data)
	b := &Block{}
	var err error

	// 1. Read Header (Order MUST MATCH SerializeAQX)
	if b.Index, err = d.UInt64(); err != nil {
		return nil, err
	}
	if b.View, err = d.UInt64(); err != nil {
		return nil, err
	}
	if b.Timestamp, err = d.UInt64(); err != nil {
		return nil, err
	}
	if b.PreviousHash, err = d.BytesArray(); err != nil {
		return nil, err
	}
	if b.MerkleRoot, err = d.BytesArray(); err != nil {
		return nil, err
	}
	if b.Hash, err = d.BytesArray(); err != nil {
		return nil, err
	}
	if b.ValidatorID, err = d.UInt64(); err != nil {
		return nil, err
	}
	if b.Signature, err = d.BytesArray(); err != nil {
		return nil, err
	}

	// 2. Read Transactions List (Hashes Only)
	txCount, err := d.UInt32()
	if err != nil {
		return nil, err
	}

	b.TransactionHashes = make([][]byte, 0, txCount)
	for i := uint32(0); i < txCount; i++ {
		hash, err := d.FixedBytes(32)
		if err != nil {
			return nil, err
		}
		
		// Copy to avoid zero-copy memory faults
		safeHash := make([]byte, 32)
		copy(safeHash, hash)

		b.TransactionHashes = append(b.TransactionHashes, safeHash)
	}

	return b, nil
}
