package block

import (
	"sync/atomic"
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
)

// Global atomic counter specifically for the test environment to guarantee
// that tests running in parallel don't accidentally reuse the same nonce
var testNonceCounter uint64 = 0

func createTestTx(t *testing.T, node *identity.NodeIdentity) *transaction.Transaction {
	// Increment the global test nonce safely
	nonce := atomic.AddUint64(&testNonceCounter, 1)

	// Initialize the transaction (defaults Nonce to 0)
	tx := transaction.NewTransaction(node, [32]byte{}, "AQX Block Payload")

	// Manually inject the incremented nonce before signing
	tx.Nonce = nonce

	if err := tx.SignWithIdentity(node); err != nil {
		t.Fatal(err)
	}
	return tx
}

func TestBlockFinalizeAndVerify(t *testing.T) {

	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	tx := createTestTx(t, node)

	block := NewBlock(
		1, // height
		0, // view
		[]byte("prev_hash"),
		[]*transaction.Transaction{tx},
	)

	if err := block.Finalize(node); err != nil {
		t.Fatal(err)
	}

	valid, err := block.Verify(signer, node.PublicKey)
	if err != nil || !valid {
		t.Fatal("Block verification failed")
	}
}

func TestBlockTamperFails(t *testing.T) {

	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	tx := createTestTx(t, node)

	block := NewBlock(
		1,
		0,
		[]byte("prev_hash"),
		[]*transaction.Transaction{tx},
	)

	if err := block.Finalize(node); err != nil {
		t.Fatal(err)
	}

	// Tamper block
	block.Hash = []byte("fake_hash")

	valid, _ := block.Verify(signer, node.PublicKey)
	if valid {
		t.Fatal("Tampered block should fail")
	}
}
