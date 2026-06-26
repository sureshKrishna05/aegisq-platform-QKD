package ledger

import (
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/block"
	"github.com/sureshKrishna05/aegisq-framework/core/consensus"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
)

func TestFullChainWithDilithium(t *testing.T) {

	signer, err := crypto.NewDilithiumSigner()
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()

	// Create validator identity
	node, err := identity.NewNodeIdentity("validator-1", 1, signer)
	if err != nil {
		t.Fatal(err)
	}

	// Create validator set
	vs := consensus.NewValidatorSet()
	vs.AddValidator(1, node.PublicKey)

	// --- Genesis block ---
	genesisTx := transaction.NewTransaction(
		node,
		[32]byte{},
		"genesis_data",
	)

	if err := genesisTx.SignWithIdentity(node); err != nil {
		t.Fatal(err)
	}

	genesis := block.NewBlock(uint64(
		0), uint64(// height
		0), // view
		[]byte("genesis"),
		[]*transaction.Transaction{genesisTx},
	)

	if err := genesis.Finalize(node); err != nil {
		t.Fatal(err)
	}

	ledger := NewLedger(genesis, vs)

	// --- Next block ---
	tx := transaction.NewTransaction(
		node,
		[32]byte{},
		"block_data",
	)
	tx.Nonce = 1

	if err := tx.SignWithIdentity(node); err != nil {
		t.Fatal(err)
	}

	newBlock := block.NewBlock(uint64(
		1), uint64(// height
		0),            // view
		genesis.Hash, // prev hash
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
