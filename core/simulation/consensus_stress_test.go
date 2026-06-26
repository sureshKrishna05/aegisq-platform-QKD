package simulation

import (
	"fmt"
	"os"
	"testing"

	"github.com/sureshKrishna05/aegisq-framework/core/block"
	"github.com/sureshKrishna05/aegisq-framework/core/crypto"
	"github.com/sureshKrishna05/aegisq-framework/core/identity"
	"github.com/sureshKrishna05/aegisq-framework/core/storage"
)

// BenchmarkStorageThroughput tests PebbleDB performance under extreme synthetic loads
func BenchmarkStorageThroughput(b *testing.B) {
	// Clean up any previous test db
	os.RemoveAll("test_bench.db")
	defer os.RemoveAll("test_bench.db")

	rawDB, err := storage.Open("test_bench.db", nil)
	if err != nil {
		b.Fatal(err)
	}
	defer rawDB.Close()

	// Use LRU Cache
	db := storage.NewCachedStore(rawDB, 1000)

	signer, _ := crypto.NewDilithiumSigner()
	leader, _ := identity.NewNodeIdentity("bench-leader", 1, signer)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		
		// Generate 1000 transactions per block
		txs, _ := GenerateSyntheticDataset(1000, leader)
		
		prevHash := []byte(fmt.Sprintf("prev_hash_%d", i))
		newBlock := block.NewBlock(uint64(i+1), 0, prevHash, txs)
		newBlock.Finalize(leader)

		b.StartTimer()
		
		// Measure just the atomic database save
		err := db.SaveBlock(newBlock)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCacheHit tests the performance of the LRU cache
func BenchmarkCacheHit(b *testing.B) {
	os.RemoveAll("test_cache.db")
	defer os.RemoveAll("test_cache.db")

	rawDB, err := storage.Open("test_cache.db", nil)
	if err != nil {
		b.Fatal(err)
	}
	defer rawDB.Close()

	db := storage.NewCachedStore(rawDB, 10) // Cache size 10

	signer, _ := crypto.NewDilithiumSigner()
	leader, _ := identity.NewNodeIdentity("cache-leader", 1, signer)

	// Pre-load 5 blocks
	for i := 1; i <= 5; i++ {
		txs, _ := GenerateSyntheticDataset(10, leader)
		newBlock := block.NewBlock(uint64(i), 0, []byte("prev"), txs)
		newBlock.Finalize(leader)
		db.SaveBlock(newBlock)
	}

	b.ResetTimer()
	b.ReportAllocs()

	// This will exclusively hit the LRU RAM Cache (Zero Disk IO)
	for i := 0; i < b.N; i++ {
		height := uint64((i % 5) + 1)
		_, err := db.GetBlock(height)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParallelVerification tests the multi-core scaling of signature verification
func BenchmarkParallelVerification(b *testing.B) {
	signer, _ := crypto.NewDilithiumSigner()
	leader, _ := identity.NewNodeIdentity("parallel-leader", 1, signer)

	// Pre-generate 1 block with 5000 transactions
	txs, _ := GenerateSyntheticDataset(5000, leader)
	testBlock := block.NewBlock(1, 0, []byte("prev"), txs)
	testBlock.Finalize(leader)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid, err := testBlock.Verify(signer, leader.PublicKey)
		if err != nil || !valid {
			b.Fatalf("Verification failed: %v", err)
		}
	}
}
