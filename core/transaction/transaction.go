package transaction

import (
	"errors"
	"time"

	"github.com/sureshKrishna05/aegisq-framework/core/aqx"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
)

type Transaction struct {
	SenderID  string
	PublicKey []byte
	Algorithm string
	DataHash  [32]byte
	Metadata  string
	Timestamp uint64
	Nonce     uint64
	Signature []byte
}

func NewTransaction(node *identity.NodeIdentity, dataHash [32]byte, metadata string) *Transaction {
	return &Transaction{
		SenderID:  node.NodeID,
		PublicKey: node.PublicKey,
		Algorithm: node.Algorithm(),
		DataHash:  dataHash,
		Metadata:  metadata,
		Timestamp: uint64(time.Now().Unix()),
		Nonce:     0, // Initialize nonce
	}
}

// computePayloadHash uses the AQX Encoder to deterministically pack the fields
// WITHOUT the signature, guaranteeing a perfectly stable hash for Dilithium to sign.
func (tx *Transaction) computePayloadHash() ([]byte, error) {

	if tx.SenderID == "" {
		return nil, errors.New("invalid transaction fields")
	}

	// 1. Grab a zero-allocation buffer from the AQX pool
	e := aqx.AcquireEncoder()
	defer e.Release()

	// 2. Deterministically encode according to AQX RFC rules
	// Encoding Order MUST be: SenderID, PublicKey, Algorithm, DataHash, Nonce, Metadata, Timestamp
	e.String(tx.SenderID)
	e.BytesArray(tx.PublicKey)
	e.String(tx.Algorithm)
	e.FixedBytes(tx.DataHash[:])
	e.UInt64(tx.Nonce)
	e.String(tx.Metadata)
	e.UInt64(tx.Timestamp)

	// 3. Hash the canonical AQX bytes.
	// The buffer is safely returned to the pool after crypto.Hash returns!
	return crypto.Hash(e.Bytes()), nil
}

func (tx *Transaction) Hash() ([]byte, error) {
	return tx.computePayloadHash()
}

// SignWithIdentity triggers the cryptographic pipeline: AQX -> Hash -> Dilithium
func (tx *Transaction) SignWithIdentity(node *identity.NodeIdentity) error {
	hash, err := tx.computePayloadHash()
	if err != nil {
		return err
	}

	signature, err := node.Sign(hash)
	if err != nil {
		return err
	}

	tx.Signature = signature
	return nil
}

// 🔥 REQUIRED for batch signing
func (tx *Transaction) SetSignature(sig []byte) {
	tx.Signature = sig
}

func (tx *Transaction) Verify(signer crypto.Signer) (bool, error) {
	if tx.Algorithm != signer.Algorithm() {
		return false, errors.New("algorithm mismatch")
	}

	hash, err := tx.computePayloadHash()
	if err != nil {
		return false, err
	}

	return signer.Verify(tx.PublicKey, hash, tx.Signature), nil
}

// =====================================================================
// AQX NETWORK & STORAGE LAYER
// =====================================================================

// EncodeAQX writes the transaction directly into an existing encoder buffer.
// This is essential for zero-allocation batch database writes.
func (tx *Transaction) EncodeAQX(e *aqx.Encoder) {
	e.String(tx.SenderID)
	e.BytesArray(tx.PublicKey)
	e.String(tx.Algorithm)
	e.FixedBytes(tx.DataHash[:])
	e.UInt64(tx.Nonce)
	e.String(tx.Metadata)
	e.UInt64(tx.Timestamp)
	e.BytesArray(tx.Signature)
}

// SerializeAQX completely flattens the Transaction (INCLUDING the signature)
// into a final, portable byte array suitable for network transport.
func (tx *Transaction) SerializeAQX() []byte {
	e := aqx.AcquireEncoder()
	defer e.Release()

	tx.EncodeAQX(e)

	// We make a copy of the final bytes so the memory pool buffer isn't
	// overwritten while this payload is moving across the network.
	out := make([]byte, len(e.Bytes()))
	copy(out, e.Bytes())
	return out
}

// DeserializeAQX acts as a constructor that reads raw AQX network bytes
// and re-inflates them back into a live Transaction object.
func DeserializeAQX(data []byte) (*Transaction, error) {
	d := aqx.NewDecoder(data)
	tx := &Transaction{}
	var err error

	// Read fields in the EXACT SAME strict order as Encode
	if tx.SenderID, err = d.String(); err != nil {
		return nil, err
	}
	if tx.PublicKey, err = d.BytesArray(); err != nil {
		return nil, err
	}
	if tx.Algorithm, err = d.String(); err != nil {
		return nil, err
	}
	
	// Read 32 fixed bytes for DataHash
	hashBytes, err := d.FixedBytes(32)
	if err != nil {
		return nil, err
	}
	copy(tx.DataHash[:], hashBytes)

	if tx.Nonce, err = d.UInt64(); err != nil {
		return nil, err
	}
	if tx.Metadata, err = d.String(); err != nil {
		return nil, err
	}
	if tx.Timestamp, err = d.UInt64(); err != nil {
		return nil, err
	}
	if tx.Signature, err = d.BytesArray(); err != nil {
		return nil, err
	}

	return tx, nil
}
