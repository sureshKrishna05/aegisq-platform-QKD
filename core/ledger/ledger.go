package ledger

import (
	"errors"
	"sync"

	"github.com/sureshKrishna05/aegisq-framework/core/block"
	"github.com/sureshKrishna05/aegisq-framework/core/consensus"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/scheduler"
)

type Ledger struct {
	Blocks       []*block.Block
	mu           sync.RWMutex // Ensuring thread safety for concurrent access
	ValidatorSet *consensus.ValidatorSet
	Scheduler    *scheduler.RoundRobinScheduler

	// NEW: State tracking to prevent Replay Attacks
	// Maps SenderID -> Last Used Nonce
	NonceTracker map[string]uint64
}

func NewLedger(genesis *block.Block, vs *consensus.ValidatorSet) *Ledger {
	s := scheduler.NewRoundRobinScheduler(vs)

	ledger := &Ledger{
		Blocks:       []*block.Block{genesis},
		ValidatorSet: vs,
		Scheduler:    s,
		NonceTracker: make(map[string]uint64), // Initialize the tracker
	}

	// FIX: Seed the NonceTracker with the genesis block's transactions!
	// If we don't do this, the Ledger thinks everyone's nonce is 0,
	// which will falsely flag valid transactions in Block 1 as Replay Attacks.
	if genesis != nil {
		for _, tx := range genesis.Transactions {
			ledger.NonceTracker[tx.SenderID] = tx.Nonce
		}
	}

	return ledger
}

func (l *Ledger) GetLastBlock() *block.Block {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Blocks[len(l.Blocks)-1]
}

func (l *Ledger) AddBlock(
	b *block.Block,
	signer crypto.Signer,
	validatorPubKey []byte,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	last := l.Blocks[len(l.Blocks)-1]

	if b.Index != last.Index+1 {
		return errors.New("invalid block index")
	}

	if string(b.PreviousHash) != string(last.Hash) {
		return errors.New("invalid previous hash linkage")
	}

	// Layer 8 leader enforcement
	expectedLeader, err := l.Scheduler.GetLeader(uint64(b.Index), uint64(b.View))
	if err != nil {
		return err
	}

	if b.ValidatorID != expectedLeader {
		return errors.New("block produced by wrong scheduled validator")
	}

	if !l.ValidatorSet.IsAuthorized(b.ValidatorID, validatorPubKey) {
		return errors.New("validator not authorized")
	}

	valid, err := b.Verify(signer, validatorPubKey)
	if err != nil || !valid {
		return errors.New("block verification failed")
	}

	for _, existing := range l.Blocks {
		if string(existing.Hash) == string(b.Hash) {
			return errors.New("duplicate block detected")
		}
	}

	// FIX: Replay Attack Defense Logic (Multi-Transaction Support)
	// We use a temporary tracker to correctly validate multiple transactions
	// from the same sender within this single block sequentially.
	tempTracker := make(map[string]uint64)
	for k, v := range l.NonceTracker {
		tempTracker[k] = v
	}

	for _, tx := range b.Transactions {
		expectedNonce := tempTracker[tx.SenderID] + 1

		if tx.Nonce != expectedNonce {
			// If we see a Nonce out of order (or repeated), reject the ENTIRE block.
			// This forces validators to only propose valid, non-replayed transactions.
			return errors.New("REPLAY ATTACK DETECTED: Invalid transaction nonce")
		}

		// Update the temp tracker for the next transaction in this loop
		tempTracker[tx.SenderID] = tx.Nonce
	}

	// If all transactions are valid, commit the temporary state to the Ledger's actual state tracker
	for k, v := range tempTracker {
		l.NonceTracker[k] = v
	}

	l.Blocks = append(l.Blocks, b)

	return nil
}

func (l *Ledger) ValidateChain(
	signer crypto.Signer,
) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i := 1; i < len(l.Blocks); i++ {

		current := l.Blocks[i]
		prev := l.Blocks[i-1]

		if current.Index != prev.Index+1 {
			return errors.New("chain index broken")
		}

		if string(current.PreviousHash) != string(prev.Hash) {
			return errors.New("chain previous hash broken")
		}

		expectedLeader, err := l.Scheduler.GetLeader(uint64(current.Index), uint64(current.View))
		if err != nil {
			return err
		}

		if current.ValidatorID != expectedLeader {
			return errors.New("invalid leader at block height")
		}

		validatorKey, exists := l.ValidatorSet.GetValidator(current.ValidatorID)
		if !exists {
			return errors.New("block signed by unknown validator")
		}

		valid, err := current.Verify(signer, validatorKey)
		if err != nil || !valid {
			return errors.New("block verification failed")
		}
	}

	return nil
}
