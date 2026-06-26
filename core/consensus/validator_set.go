package consensus

import (
	"sort"
)

type ValidatorSet struct {
	validators map[uint64][]byte
}

func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{
		validators: make(map[uint64][]byte),
	}
}

func (v *ValidatorSet) AddValidator(validatorID uint64, publicKey []byte) {
	v.validators[validatorID] = publicKey
}

func (v *ValidatorSet) RemoveValidator(validatorID uint64) {
	delete(v.validators, validatorID)
}

func (v *ValidatorSet) IsAuthorized(validatorID uint64, publicKey []byte) bool {
	registeredKey, exists := v.validators[validatorID]
	if !exists {
		return false
	}
	if string(registeredKey) != string(publicKey) {
		return false
	}
	return true
}

func (v *ValidatorSet) GetValidator(validatorID uint64) ([]byte, bool) {
	key, exists := v.validators[validatorID]
	return key, exists
}

func (v *ValidatorSet) Count() int {
	return len(v.validators)
}

func (v *ValidatorSet) GetValidatorIDs() []uint64 {
	ids := make([]uint64, 0, len(v.validators))
	for id := range v.validators {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
