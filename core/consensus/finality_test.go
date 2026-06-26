package consensus

import "testing"

func TestFinalityFlow(t *testing.T) {

	vs := NewValidatorSet()
	vs.AddValidator(1, []byte("k1"))
	vs.AddValidator(2, []byte("k2"))
	vs.AddValidator(3, []byte("k3"))
	vs.AddValidator(4, []byte("k4"))

	vp := NewVotePool(vs)
	fe := NewFinalityEngine(vp)

	height := 1
	view := 0

	_ = vp.AddVote(Vote{ValidatorID: 1, BlockHash: [32]byte{}, View: uint64(view), Type: Prepare, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 2, BlockHash: [32]byte{}, View: uint64(view), Type: Prepare, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 3, BlockHash: [32]byte{}, View: uint64(view), Type: Prepare, Signature: nil})

	if !fe.TryPrepare(uint64(height), [32]byte{}, uint64(view)) {
		t.Fatal("Prepare should succeed")
	}

	_ = vp.AddVote(Vote{ValidatorID: 1, BlockHash: [32]byte{}, View: uint64(view), Type: Commit, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 2, BlockHash: [32]byte{}, View: uint64(view), Type: Commit, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 3, BlockHash: [32]byte{}, View: uint64(view), Type: Commit, Signature: nil})

	if err := fe.TryCommit(uint64(height), [32]byte{}, uint64(view)); err != nil {
		t.Fatal("Commit should succeed:", err)
	}

	if !fe.IsFinalized(uint64(height), [32]byte{}) {
		t.Fatal("Block should be finalized")
	}
}

func TestForkPrevention(t *testing.T) {

	vs := NewValidatorSet()
	vs.AddValidator(1, []byte("k1"))
	vs.AddValidator(2, []byte("k2"))
	vs.AddValidator(3, []byte("k3"))
	vs.AddValidator(4, []byte("k4"))

	vp := NewVotePool(vs)
	fe := NewFinalityEngine(vp)

	height := 1
	view := 0

	_ = vp.AddVote(Vote{ValidatorID: 1, BlockHash: [32]byte{}, View: uint64(view), Type: Prepare, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 2, BlockHash: [32]byte{}, View: uint64(view), Type: Prepare, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 3, BlockHash: [32]byte{}, View: uint64(view), Type: Prepare, Signature: nil})
	fe.TryPrepare(uint64(height), [32]byte{}, uint64(view))

	_ = vp.AddVote(Vote{ValidatorID: 1, BlockHash: [32]byte{}, View: uint64(view), Type: Commit, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 2, BlockHash: [32]byte{}, View: uint64(view), Type: Commit, Signature: nil})
	_ = vp.AddVote(Vote{ValidatorID: 3, BlockHash: [32]byte{}, View: uint64(view), Type: Commit, Signature: nil})
	_ = fe.TryCommit(uint64(height), [32]byte{}, uint64(view))

	err := fe.TryCommit(uint64(height), [32]byte{1}, uint64(view))
	if err == nil {
		t.Fatal("Fork should not be allowed")
	}
}
