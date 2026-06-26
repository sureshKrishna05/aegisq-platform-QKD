package scheduler

import (
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/consensus"
)

func TestRoundRobinRotation(t *testing.T) {

	vs := consensus.NewValidatorSet()

	vs.AddValidator(1, []byte("a"))
	vs.AddValidator(2, []byte("b"))
	vs.AddValidator(3, []byte("c"))

	s := NewRoundRobinScheduler(vs)

	tests := []struct {
		height   int
		view     int
		expected uint64
	}{
		{0, 0, 1},
		{1, 0, 2},
		{2, 0, 3},
		{3, 0, 1},
		{1, 1, 3}, // failover
		{1, 2, 1}, // next failover
	}

	for _, test := range tests {

		leader, err := s.GetLeader(uint64(test.height), uint64(test.view))
		if err != nil {
			t.Fatal(err)
		}

		if leader != test.expected {
			t.Fatalf("expected %d, got %d", test.expected, leader)
		}
	}
}
