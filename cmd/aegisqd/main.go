package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sureshKrishna05/aegisq-framework/core/block"
	"github.com/sureshKrishna05/aegisq-framework/core/consensus"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/event"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/network/qkd"
	"github.com/sureshKrishna05/aegisq-framework/core/scheduler"
	"github.com/sureshKrishna05/aegisq-framework/core/simulation"
	"github.com/sureshKrishna05/aegisq-framework/core/storage"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
)

type ConsoleSubscriber struct{}

func (ConsoleSubscriber) Name() string { return "ConsoleSubscriber" }
func (ConsoleSubscriber) InterestedIn() []event.EventType {
	return []event.EventType{
		event.BlockPersisted,
		event.IntegrityCheckPassed,
		event.BlocksPruned,
		event.SnapshotCreated,
	}
}
func (ConsoleSubscriber) Handle(e event.Event) error {
	fmt.Printf("⚡ [EVENT BUS] %s published by %s | Payload: %v\n", e.Type, e.Source, e.Payload)
	return nil
}

func main() {

	if len(os.Args) == 4 && os.Args[1] == "gettx" {
		height, _ := strconv.Atoi(os.Args[2])
		index, _ := strconv.Atoi(os.Args[3])

		db, err := storage.Open("aegisq.db", nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		blockObj, err := db.GetBlock(uint64(height))
		if err != nil {
			log.Fatal(err)
		}

		if index >= len(blockObj.Transactions) {
			log.Fatal("transaction index out of range")
		}

		tx := blockObj.Transactions[index]
		printTxDetails(int(blockObj.Index), index, tx)
		return
	}

	if len(os.Args) == 3 && os.Args[1] == "gettxhash" {
		hash := os.Args[2]

		db, err := storage.Open("aegisq.db", nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		blockObj, index, err := db.GetTransactionByHash(hash)
		if err != nil {
			log.Fatal(err)
		}

		tx := blockObj.Transactions[index]
		printTxDetails(int(blockObj.Index), index, tx)
		return
	}

	signer, err := crypto.NewDilithiumSigner()
	if err != nil {
		panic(err)
	}

	var validators []*identity.NodeIdentity
	for i := 1; i <= 4; i++ {
		node, err := identity.NewNodeIdentity(
			fmt.Sprintf("validator-%d", i),
			uint64(i),
			signer,
		)
		if err != nil {
			log.Fatal(err)
		}
		validators = append(validators, node)
	}

	fmt.Println("Validators initialized.")

	// NEW: Initialize QKD Engine and generate keys
	fmt.Println("Initializing QKD Engine (BB84)...")
	// Using relative path assuming we run from project root
	qkdEngine := qkd.NewEngine("qkd_engine/bb84_sim.py")

	secureChannels := make(map[uint64]*qkd.SecureChannel)
	for _, v := range validators {
		fmt.Printf("Establishing Quantum Session Key for Validator %d...\n", v.ValidatorID)
		res, err := qkdEngine.GenerateSessionKey(1024, false, 0.0)
		if err != nil {
			log.Fatalf("QKD Failed: %v", err)
		}
		channel, err := qkd.NewSecureChannelFromHex(res.SymmetricKeyHex)
		if err != nil {
			log.Fatalf("Secure channel failed: %v", err)
		}
		secureChannels[v.ValidatorID] = channel
	}
	fmt.Println("All secure channels established.")

	vs := consensus.NewValidatorSet()
	for _, v := range validators {
		vs.AddValidator(v.ValidatorID, v.PublicKey)
	}

	sched := scheduler.NewRoundRobinScheduler(vs)

	// Phase 4: AES Event Bus Integration
	bus := event.NewEventBus()
	bus.Subscribe(ConsoleSubscriber{})

	rawDB, err := storage.Open("aegisq.db", bus)
	if err != nil {
		log.Fatal(err)
	}
	defer rawDB.Close()

	// Wrap PebbleDB with an LRU cache (Capacity: 1000 blocks)
	db := storage.NewCachedStore(rawDB, 1000)

	fmt.Println("Running Crash Recovery & Integrity Verification...")
	if err := db.CheckIntegrity(); err != nil {
		log.Fatal("Database integrity check failed:", err)
	}
	fmt.Println("Database integrity verified.")

	height, err := db.GetLatestHeight()
	if err != nil {
		log.Fatal(err)
	}

	var previousHash []byte
	if height > 0 {
		lastBlock, err := db.GetBlock(height)
		if err != nil {
			log.Fatal(err)
		}
		previousHash = lastBlock.Hash
		fmt.Println("Restored height:", height)
	} else {
		fmt.Println("No chain found. Starting fresh.")
	}

	view := uint64(0)
	leaderID, err := sched.GetLeader(height+1, view)
	if err != nil {
		log.Fatal(err)
	}

	var leader *identity.NodeIdentity
	for _, v := range validators {
		if v.ValidatorID == leaderID {
			leader = v
			break
		}
	}

	if leader == nil {
		log.Fatal("leader not found")
	}

	fmt.Println("Leader selected:", leader.NodeID)

	startTx := time.Now()
	txs, err := simulation.GenerateSyntheticDataset(10000, leader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generated synthetic storage transactions:", len(txs))
	fmt.Println("Transaction generation time:", time.Since(startTx))

	startFinalize := time.Now()
	newBlock := block.NewBlock(
		height+1,
		view,
		previousHash,
		txs,
	)

	if err := newBlock.Finalize(leader); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Block finalize time:", time.Since(startFinalize))
	fmt.Println("Proposed block height:", newBlock.Index)

	var blockHashArray [32]byte
	copy(blockHashArray[:], newBlock.Hash)

	votePool := consensus.NewVotePool(vs)
	for _, v := range validators {
		vote := consensus.Vote{
			ValidatorID: v.ValidatorID,
			BlockHash:   blockHashArray,
			View:        view,
			Type:        consensus.Prepare,
		}

		// QKD ENCRYPTION LAYER
		channel := secureChannels[v.ValidatorID]
		ciphertext, err := channel.Encrypt(vote.SerializeAQX())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[QKD NETWORK] Encrypted PREPARE vote from Validator %d. Ciphertext (first 16 bytes): %x...\n", v.ValidatorID, ciphertext[:16])

		// QKD DECRYPTION LAYER
		plaintext, err := channel.Decrypt(ciphertext)
		if err != nil {
			log.Fatal(err)
		}
		decryptedVote, err := consensus.DeserializeAQX(plaintext)
		if err != nil {
			log.Fatal(err)
		}

		_ = votePool.AddVote(*decryptedVote)
	}

	if !votePool.HasQuorum(blockHashArray, view, consensus.Prepare) {
		fmt.Println("Prepare quorum NOT reached.")
		return
	}
	fmt.Println("Prepare quorum reached.")

	for _, v := range validators {
		vote := consensus.Vote{
			ValidatorID: v.ValidatorID,
			BlockHash:   blockHashArray,
			View:        view,
			Type:        consensus.Commit,
		}

		// QKD ENCRYPTION LAYER
		channel := secureChannels[v.ValidatorID]
		ciphertext, err := channel.Encrypt(vote.SerializeAQX())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[QKD NETWORK] Encrypted COMMIT vote from Validator %d. Ciphertext (first 16 bytes): %x...\n", v.ValidatorID, ciphertext[:16])

		// QKD DECRYPTION LAYER
		plaintext, err := channel.Decrypt(ciphertext)
		if err != nil {
			log.Fatal(err)
		}
		decryptedVote, err := consensus.DeserializeAQX(plaintext)
		if err != nil {
			log.Fatal(err)
		}

		_ = votePool.AddVote(*decryptedVote)
	}

	if !votePool.HasQuorum(blockHashArray, view, consensus.Commit) {
		fmt.Println("Commit quorum NOT reached.")
		return
	}
	fmt.Println("Commit quorum reached.")

	if err := db.SaveBlock(newBlock); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Block committed at height:", newBlock.Index)
	printBlockSummary(newBlock)

	startServer(db)
}

func printTxDetails(height int, index int, tx *transaction.Transaction) {
	fmt.Println("----- Transaction Details -----")
	fmt.Println("Block Height:", height)
	fmt.Println("Transaction Index:", index)
	fmt.Println("Sender:", tx.SenderID)
	fmt.Println("Algorithm:", tx.Algorithm)
	fmt.Println("DataHash:", tx.DataHash)
	fmt.Println("Metadata:", tx.Metadata)
	fmt.Println("Timestamp:", tx.Timestamp)
	fmt.Printf("Signature: %x\n", tx.Signature)
	fmt.Println("--------------------------------")
}

func printBlockSummary(b *block.Block) {
	fmt.Println("\n========= BLOCK SUMMARY =========")
	fmt.Println("Height:", b.Index)
	fmt.Printf("Hash: %x\n", b.Hash)
	fmt.Printf("Previous: %x\n", b.PreviousHash)
	fmt.Println("Total Transactions:", len(b.Transactions))
	for i := 0; i < 5 && i < len(b.Transactions); i++ {
		tx := b.Transactions[i]
		fmt.Println("  Tx", i+1)
		fmt.Println("   Sender:", tx.SenderID)
		fmt.Println("   DataHash:", tx.DataHash)
	}
	if len(b.Transactions) > 5 {
		fmt.Println("  ...")
	}
}


