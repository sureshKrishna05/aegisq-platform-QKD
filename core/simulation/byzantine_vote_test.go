package simulation

import (
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/consensus"
)

func TestByzantineEquivocationAttack(t *testing.T) {

	vs := consensus.NewValidatorSet()

	vs.AddValidator(1, []byte("k1"))
	vs.AddValidator(2, []byte("k2"))
	vs.AddValidator(3, []byte("k3"))
	vs.AddValidator(4, []byte("k4"))

	vp := consensus.NewVotePool(vs)

	view := uint64(1)

	vp.AddVote(consensus.Vote{
		ValidatorID: 1,
		BlockHash:   [32]byte{1},
		View:        view,
		Type:        consensus.Prepare,
	})

	vp.AddVote(consensus.Vote{
		ValidatorID: 2,
		BlockHash:   [32]byte{1},
		View:        view,
		Type:        consensus.Prepare,
	})

	vp.AddVote(consensus.Vote{
		ValidatorID: 3,
		BlockHash:   [32]byte{1},
		View:        view,
		Type:        consensus.Prepare,
	})

	vp.AddVote(consensus.Vote{
		ValidatorID: 3,
		BlockHash:   [32]byte{2},
		View:        view,
		Type:        consensus.Prepare,
	})

	vp.AddVote(consensus.Vote{
		ValidatorID: 4,
		BlockHash:   [32]byte{1},
		View:        view,
		Type:        consensus.Prepare,
	})

	if !vp.HasQuorum([32]byte{1}, view, consensus.Prepare) {
		t.Fatal("Expected quorum for blockA not reached")
	}

	if vp.HasQuorum([32]byte{2}, view, consensus.Prepare) {
		t.Fatal("Byzantine equivocation formed illegal quorum for blockB")
	}
}
