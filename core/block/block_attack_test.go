package block

import (
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
)

func TestAttack_ModifyTransactionInsideBlock(t *testing.T) {

	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	tx := transaction.NewTransaction(node, [32]byte{}, "data")
	tx.SignWithIdentity(node)

	block := NewBlock(
		1,
		0,
		[]byte("prev_hash"),
		[]*transaction.Transaction{tx},
	)

	block.Finalize(node)

	// Corrupt transaction signature after block finalized
	block.Transactions[0].Signature = []byte("corrupted")

	valid, _ := block.Verify(signer, node.PublicKey)

	if valid {
		t.Fatal("Attack succeeded: block should be invalid after tx mutation")
	}
}

func TestAttack_ModifyBlockHash(t *testing.T) {

	signer := &crypto.Ed25519Signer{}
	node, _ := identity.NewNodeIdentity("validator-1", 1, signer)

	tx := transaction.NewTransaction(node, [32]byte{}, "data")
	tx.SignWithIdentity(node)

	block := NewBlock(
		1,
		0,
		[]byte("prev_hash"),
		[]*transaction.Transaction{tx},
	)

	block.Finalize(node)

	block.Hash = []byte("corrupted")

	valid, _ := block.Verify(signer, node.PublicKey)

	if valid {
		t.Fatal("Corrupted block hash should fail")
	}
}
