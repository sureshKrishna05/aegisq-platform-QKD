package scheduler

import (
	"errors"
	"github.com/sureshKrishna05/aegisq-framework/core/consensus"
)

type RoundRobinScheduler struct {
	orderedValidators []uint64
}

func NewRoundRobinScheduler(vs *consensus.ValidatorSet) *RoundRobinScheduler {
	return &RoundRobinScheduler{
		orderedValidators: vs.GetValidatorIDs(),
	}
}

func (r *RoundRobinScheduler) GetLeader(height uint64, view uint64) (uint64, error) {
	if len(r.orderedValidators) == 0 {
		return 0, errors.New("no validators registered")
	}
	pos := (height + view) % uint64(len(r.orderedValidators))
	return r.orderedValidators[pos], nil
}
