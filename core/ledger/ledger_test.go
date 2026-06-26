package ledger

import (
	"sync/atomic"
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/block"
	"github.com/sureshKrishna05/aegisq-framework/core/consensus"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
)

// Global atomic counter specifically for the test environment to guarantee
// that tests running in parallel don't accidentally reuse the same nonce
var testNonceCounter uint64 = 0

func setupLedger(t *testing.T) (*Ledger, *identity.NodeIdentity, crypto.Signer) {

	signer := &crypto.Ed25519Signer{}

	node, err := identity.NewNodeIdentity("validator-1", 1, signer)
	if err != nil {
		t.Fatal(err)
	}

	vs := consensus.NewValidatorSet()
	vs.AddValidator(1, node.PublicKey)

	// Genesis must contain at least one transaction
	tx := createDummyTransaction(t, node)

	genesis := block.NewBlock(uint64(
		0), uint64(// height
		0), // view
		[]byte("genesis"),
		[]*transaction.Transaction{tx},
	)

	if err := genesis.Finalize(node); err != nil {
		t.Fatal(err)
	}

	ledger := NewLedger(genesis, vs)

	return ledger, node, signer
}

// FIX: Helper function now injects an incrementing Nonce into every transaction
// generated across the entire test suite to prevent Replay Attack rejections.
func createDummyTransaction(t *testing.T, node *identity.NodeIdentity) *transaction.Transaction {

	// Increment the global test nonce safely
	nonce := atomic.AddUint64(&testNonceCounter, 1)

	// Call the constructor with 3 arguments as expected
	tx := transaction.NewTransaction(
		node,
		[32]byte{},
		"dummy_data",
	)

	// Manually inject the incremented nonce before signing
	tx.Nonce = nonce

	err := tx.SignWithIdentity(node)
	if err != nil {
		t.Fatal(err)
	}

	return tx
}

func TestLedgerAddBlock(t *testing.T) {

	ledger, node, signer := setupLedger(t)

	tx := createDummyTransaction(t, node)

	newBlock := block.NewBlock(uint64(
		1), uint64(// height
		0), // view
		ledger.GetLastBlock().Hash,
		[]*transaction.Transaction{tx},
	)

	if err := newBlock.Finalize(node); err != nil {
		t.Fatal(err)
	}

	if err := ledger.AddBlock(newBlock, signer, node.PublicKey); err != nil {
		t.Fatal("Failed to add valid block:", err)
	}
}

func TestLedgerRejectWrongIndex(t *testing.T) {

	ledger, node, signer := setupLedger(t)

	tx := createDummyTransaction(t, node)

	newBlock := block.NewBlock(uint64(
		2), uint64(// wrong height
		0), // view
		ledger.GetLastBlock().Hash,
		[]*transaction.Transaction{tx},
	)

	if err := newBlock.Finalize(node); err != nil {
		t.Fatal(err)
	}

	if err := ledger.AddBlock(newBlock, signer, node.PublicKey); err == nil {
		t.Fatal("Block with wrong index should fail")
	}
}

func TestLedgerRejectWrongPreviousHash(t *testing.T) {

	ledger, node, signer := setupLedger(t)

	tx := createDummyTransaction(t, node)

	newBlock := block.NewBlock(uint64(
		1), uint64(0),
		[]byte("wrong_hash"),
		[]*transaction.Transaction{tx},
	)

	if err := newBlock.Finalize(node); err != nil {
		t.Fatal(err)
	}

	if err := ledger.AddBlock(newBlock, signer, node.PublicKey); err == nil {
		t.Fatal("Block with wrong previous hash should fail")
	}
}

func TestLedgerValidateChain(t *testing.T) {

	ledger, node, signer := setupLedger(t)

	tx := createDummyTransaction(t, node)

	newBlock := block.NewBlock(uint64(
		1), uint64(0),
		ledger.GetLastBlock().Hash,
		[]*transaction.Transaction{tx},
	)

	if err := newBlock.Finalize(node); err != nil {
		t.Fatal(err)
	}

	if err := ledger.AddBlock(newBlock, signer, node.PublicKey); err != nil {
		t.Fatal(err)
	}

	if err := ledger.ValidateChain(signer); err != nil {
		t.Fatal("Chain validation failed:", err)
	}
}
