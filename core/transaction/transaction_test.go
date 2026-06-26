package transaction

import (
	"bytes"
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
)

// =====================================================================
// INTEGRATION TESTS
// =====================================================================

func TestTransactionSignVerify(t *testing.T) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	// Updated to match your new 3-argument constructor
	tx := NewTransaction(node, [32]byte{}, "Test File")

	err := tx.SignWithIdentity(node)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	valid, err := tx.Verify(signer)
	if err != nil || !valid {
		t.Fatal("Transaction verification failed")
	}
}

func TestTransactionTamperFails(t *testing.T) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	tx := NewTransaction(node, [32]byte{}, "Test File")
	tx.SignWithIdentity(node)

	tx.Metadata = "HACKED"

	valid, _ := tx.Verify(signer)
	if valid {
		t.Fatal("Tampered transaction should fail")
	}
}

func TestTransactionAQXSerialization(t *testing.T) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	tx := NewTransaction(node, [32]byte{}, "AQX Integration Test")
	tx.Nonce = 42 // Increment nonce to ensure it serializes properly
	tx.SignWithIdentity(node)

	// 1. Serialize using our zero-allocation memory pool
	rawBytes := tx.SerializeAQX()

	// 2. Deserialize using our zero-copy cursor
	decodedTx, err := DeserializeAQX(rawBytes)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// 3. Verify absolute fidelity
	if decodedTx.SenderID != tx.SenderID ||
		decodedTx.DataHash != tx.DataHash ||
		decodedTx.Nonce != tx.Nonce ||
		!bytes.Equal(decodedTx.Signature, tx.Signature) {
		t.Fatal("Decoded AQX transaction does not match the original!")
	}
}

// =====================================================================
// PERFORMANCE BENCHMARKS
// =====================================================================

// Benchmark the generation of the Canonical Hash (AQX Encoding + Cryptographic Hash)
func BenchmarkTransaction_Hash(b *testing.B) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)
	tx := NewTransaction(node, [32]byte{}, "Benchmark Payload")

	b.ReportAllocs() // Will prove our zero-allocation design
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tx.Nonce = uint64(i)
		_, _ = tx.Hash()
	}
}

// Benchmark the flattening of the entire struct for network transport
func BenchmarkTransaction_SerializeAQX(b *testing.B) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)
	tx := NewTransaction(node, [32]byte{}, "Benchmark Payload")
	tx.SignWithIdentity(node)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tx.SerializeAQX()
	}
}

// Benchmark the zero-copy reconstruction of a Transaction from network bytes
func BenchmarkTransaction_DeserializeAQX(b *testing.B) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)
	tx := NewTransaction(node, [32]byte{}, "Benchmark Payload")
	tx.SignWithIdentity(node)

	rawBytes := tx.SerializeAQX()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = DeserializeAQX(rawBytes)
	}
}
