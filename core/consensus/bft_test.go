package consensus

import (
	"bytes"
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
)

// =====================================================================
// BFT INTEGRATION TESTS
// =====================================================================

func TestPrepareQuorum(t *testing.T) {
	vs := NewValidatorSet()
	vs.AddValidator(1, []byte("pub1"))
	vs.AddValidator(2, []byte("pub2"))
	vs.AddValidator(3, []byte("pub3"))
	vs.AddValidator(4, []byte("pub4"))

	vp := NewVotePool(vs)

	// FIX: Explicitly name fields to prevent InvalidStructLit errors
	_ = vp.AddVote(Vote{ValidatorID: 1, BlockHash: [32]byte{}, View: 1, Type: Prepare, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 2, BlockHash: [32]byte{}, View: 1, Type: Prepare, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 3, BlockHash: [32]byte{}, View: 1, Type: Prepare, Signature: nil})

	if !vp.HasQuorum([32]byte{}, 1, Prepare) {
		t.Fatal("Expected quorum for Prepare phase, got none")
	}
}

func TestCommitQuorum(t *testing.T) {
	vs := NewValidatorSet()
	vs.AddValidator(1, []byte("pub1"))
	vs.AddValidator(2, []byte("pub2"))
	vs.AddValidator(3, []byte("pub3"))
	vs.AddValidator(4, []byte("pub4"))

	vp := NewVotePool(vs)

	_ = vp.AddVote(Vote{ValidatorID: 1, BlockHash: [32]byte{}, View: 1, Type: Commit, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 2, BlockHash: [32]byte{}, View: 1, Type: Commit, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 3, BlockHash: [32]byte{}, View: 1, Type: Commit, Signature: nil})

	if !vp.HasQuorum([32]byte{}, 1, Commit) {
		t.Fatal("Expected quorum for Commit phase, got none")
	}
}

func TestDoubleVoteRejected(t *testing.T) {
	vs := NewValidatorSet()
	vs.AddValidator(1, []byte("pub1"))

	vp := NewVotePool(vs)

	v1 := Vote{
		ValidatorID: 1,
		BlockHash: [32]byte{},
		View:        1,
		Type:        Prepare,
		Signature:   nil,
	}

	v2 := Vote{
		ValidatorID: 1,
		BlockHash:   [32]byte{1}, // Equivocation (voting for different block in same view)
		View:        1,
		Type:        Prepare,
		Signature:   nil,
	}

	err1 := vp.AddVote(v1)
	if err1 != nil {
		t.Fatalf("First vote should be valid, got: %v", err1)
	}

	err2 := vp.AddVote(v2)
	if err2 == nil {
		t.Fatal("Second vote (equivocation) should have been rejected")
	}
}

func TestUnauthorizedValidatorRejected(t *testing.T) {
	vs := NewValidatorSet()
	vs.AddValidator(1, []byte("pub1"))

	vp := NewVotePool(vs)

	err := vp.AddVote(Vote{ValidatorID: 99, BlockHash: [32]byte{}, View: 1, Type: Prepare, Signature: nil})

	if err == nil {
		t.Fatal("Vote from unauthorized validator should have been rejected")
	}
}

// =====================================================================
// AQX INTEGRATION TESTS & BENCHMARKS
// =====================================================================

func TestVoteAQXSerialization(t *testing.T) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	vote := Vote{
		ValidatorID: 1,
		BlockHash: [32]byte{},
		View:        1,
		Type:        Prepare,
		Signature:   nil,
	}

	// Sign the vote
	err := vote.Sign(node)
	if err != nil {
		t.Fatalf("Failed to sign vote: %v", err)
	}

	rawBytes := vote.SerializeAQX()

	decodedVote, err := DeserializeAQX(rawBytes)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if decodedVote.ValidatorID != vote.ValidatorID ||
		decodedVote.BlockHash != vote.BlockHash ||
		decodedVote.View != vote.View ||
		decodedVote.Type != vote.Type ||
		!bytes.Equal(decodedVote.Signature, vote.Signature) {
		t.Fatal("Decoded AQX vote does not match the original!")
	}
}

func BenchmarkVote_SerializeAQX(b *testing.B) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	vote := Vote{
		ValidatorID: 1,
		BlockHash: [32]byte{},
		View:        1,
		Type:        Prepare,
		Signature:   nil,
	}
	_ = vote.Sign(node)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = vote.SerializeAQX()
	}
}

func BenchmarkVote_DeserializeAQX(b *testing.B) {
	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	vote := Vote{
		ValidatorID: 1,
		BlockHash: [32]byte{},
		View:        1,
		Type:        Prepare,
		Signature:   nil,
	}
	_ = vote.Sign(node)
	rawBytes := vote.SerializeAQX()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = DeserializeAQX(rawBytes)
	}
}
