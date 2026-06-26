package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/cockroachdb/pebble"
	"github.com/sureshKrishna05/aegisq-framework/core/aqx"
	"github.com/sureshKrishna05/aegisq-framework/core/block"
	"github.com/sureshKrishna05/aegisq-framework/core/event"
	"github.com/sureshKrishna05/aegisq-framework/core/transaction"
)

var (
	PrefixMeta    = []byte("m_")
	PrefixBlocks  = []byte("b_")
	PrefixHashIdx = []byte("h_")
	PrefixTxIdx   = []byte("tx_")
	PrefixTxData  = []byte("txd_")
)

type DB struct {
	conn *pebble.DB
	bus  *event.EventBus
}

func Open(path string, bus *event.EventBus) (*DB, error) {
	// Performance Engineering: Pebble Batch Tuning
	// Optimized for heavy append-only blockchain workloads
	opts := &pebble.Options{
		MemTableSize:                64 << 20, // 64 MB memtables
		MemTableStopWritesThreshold: 4,        // Stop if 4 memtables are queued (256MB)
		L0CompactionThreshold:       2,        // Quick L0 to L1 flushes
		L0StopWritesThreshold:       12,       // generous stall limit
		MaxOpenFiles:                1000,
	}

	db, err := pebble.Open(path, opts)
	if err != nil {
		return nil, err
	}
	return &DB{conn: db, bus: bus}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// CheckIntegrity performs Phase 2 Database Repair and Integrity Verification.
// It verifies that the metadata matches the actual block data, rolling back
// metadata if a crash occurred during a non-atomic flush.
func (db *DB) CheckIntegrity() error {
	latest, err := db.GetLatestHeight()
	if err != nil {
		return err
	}

	if latest == 0 {
		return nil // Genesis state, nothing to verify
	}

	// Verify the latest block is actually written and readable
	_, err = db.GetBlock(latest)
	if err == nil {
		return nil // Integrity OK
	}

	// If latest block is missing, we suffered a crash that corrupted the head metadata.
	// We walk backwards to find the highest intact block.
	var validHeight uint64
	for i := latest; i > 0; i-- {
		b, err := db.GetBlock(i)
		if err == nil {
			validHeight = i
			
			// Repair metadata
			batch := db.conn.NewBatch()
			defer batch.Close()

			_ = batch.Set(makeKey(PrefixMeta, []byte("latest_height")), uint64ToBytes(validHeight), nil)
			_ = batch.Set(makeKey(PrefixMeta, []byte("latest_hash")), b.Hash, nil)
			
			return batch.Commit(pebble.Sync)
		}
	}

		// If we reach here, the entire chain is wiped, reset metadata to 0
	batch := db.conn.NewBatch()
	defer batch.Close()
	_ = batch.Set(makeKey(PrefixMeta, []byte("latest_height")), uint64ToBytes(0), nil)
	
	if err := batch.Commit(pebble.Sync); err != nil {
		return err
	}

	if db.bus != nil {
		_ = db.bus.Publish(event.Event{
			Type:    event.IntegrityCheckPassed,
			Source:  "Storage",
			Payload: latest, // Passed or recovered height
		})
	}
	return nil
}

func uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func makeKey(prefix []byte, key []byte) []byte {
	out := make([]byte, 0, len(prefix)+len(key))
	out = append(out, prefix...)
	out = append(out, key...)
	return out
}

//
// ==============================
// SAVE BLOCK
// ==============================
//

func (db *DB) SaveBlock(b *block.Block) error {
	// Create an atomic batch
	batch := db.conn.NewBatch()
	defer batch.Close()

	hashKey := makeKey(PrefixHashIdx, b.Hash)
	
	// Prevent duplicate block
	_, closer, err := db.conn.Get(hashKey)
	if err == nil {
		closer.Close()
		return errors.New("block already exists")
	} else if err != pebble.ErrNotFound {
		return err
	}

	data := b.SerializeAQX()
	heightKeyBytes := uint64ToBytes(uint64(b.Index))
	
	// Store block by height: b_<height>
	blockKey := makeKey(PrefixBlocks, heightKeyBytes)
	if err := batch.Set(blockKey, data, nil); err != nil {
		return err
	}

	// Index block hash → height: h_<hash>
	if err := batch.Set(hashKey, heightKeyBytes, nil); err != nil {
		return err
	}

	// Grab a reusable encoder to prevent allocating a new byte slice for every transaction
	e := aqx.AcquireEncoder()
	defer e.Release()

	// Index and Store Transactions
	for i, txObj := range b.Transactions {
		txHash, _ := txObj.Hash()
		
		e.Reset()
		txObj.EncodeAQX(e)
		
		// 1. Store Full AQX Transaction Payload: txd_<hash>
		txDataKey := makeKey(PrefixTxData, txHash)
		if err := batch.Set(txDataKey, e.Bytes(), nil); err != nil {
			return err
		}

		// 2. Store Transaction Metadata Index: tx_<hash>
		txIdxKey := makeKey(PrefixTxIdx, txHash)
		indexData := struct {
			Height uint64
			Index  int
		}{
			Height: uint64(b.Index),
			Index:  i,
		}

		indexBytes, err := json.Marshal(indexData)
		if err != nil {
			return err
		}

		if err := batch.Set(txIdxKey, indexBytes, nil); err != nil {
			return err
		}
	}

	// Update metadata: m_latest_height and m_latest_hash
	if err := batch.Set(makeKey(PrefixMeta, []byte("latest_height")), heightKeyBytes, nil); err != nil {
		return err
	}

	if err := batch.Set(makeKey(PrefixMeta, []byte("latest_hash")), b.Hash, nil); err != nil {
		return err
	}

	// Commit batch to disk atomically
	if err := batch.Commit(pebble.Sync); err != nil {
		return err
	}

	if db.bus != nil {
		_ = db.bus.Publish(event.Event{
			Type:    event.BlockPersisted,
			Source:  "Storage",
			Payload: b.Index,
		})
	}
	
	return nil
}

//
// ==============================
// READ BLOCK
// ==============================
//

func (db *DB) GetLatestHeight() (uint64, error) {
	key := makeKey(PrefixMeta, []byte("latest_height"))
	val, closer, err := db.conn.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	defer closer.Close()

	return bytesToUint64(val), nil
}

func (db *DB) GetBlock(height uint64) (*block.Block, error) {
	key := makeKey(PrefixBlocks, uint64ToBytes(height))
	data, closer, err := db.conn.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, errors.New("block not found")
		}
		return nil, err
	}
	defer closer.Close()

	// CRITICAL FIX: Pebble frees `data` when `closer.Close()` is called.
	// Since AQX Decoder uses zero-copy slicing, we MUST copy the payload
	// into the heap before deserializing to prevent SIGSEGV mem faults!
	safeData := make([]byte, len(data))
	copy(safeData, data)

	b, err := block.DeserializeAQX(safeData)
	if err != nil {
		return nil, err
	}

	// Reconstruct Transactions from the separated txd_ bucket
	b.Transactions = make([]*transaction.Transaction, 0, len(b.TransactionHashes))
	for _, hash := range b.TransactionHashes {
		txDataKey := makeKey(PrefixTxData, hash)
		txData, closerTx, err := db.conn.Get(txDataKey)
		if err != nil {
			return nil, err
		}

		safeTxData := make([]byte, len(txData))
		copy(safeTxData, txData)
		closerTx.Close()

		txObj, err := transaction.DeserializeAQX(safeTxData)
		if err != nil {
			return nil, err
		}
		
		b.Transactions = append(b.Transactions, txObj)
	}

	return b, nil
}

// ==============================
// PHASE 4: MAINTENANCE & PRUNING
// ==============================

// CreateSnapshot creates a hard-linked PebbleDB checkpoint.
// This executes instantly and requires almost zero additional disk space.
func (db *DB) CreateSnapshot(destDir string) error {
	if err := db.conn.Checkpoint(destDir); err != nil {
		return err
	}
	if db.bus != nil {
		_ = db.bus.Publish(event.Event{
			Type:    event.SnapshotCreated,
			Source:  "Storage",
			Payload: destDir,
		})
	}
	return nil
}

// CompactDB forces Pebble to immediately compact all SSTables.
// Useful after massive pruning operations or initial fast-sync.
func (db *DB) CompactDB() error {
	return db.conn.Compact([]byte{0x00}, []byte{0xFF, 0xFF, 0xFF, 0xFF}, true)
}

// PruneOldBlocks securely deletes blocks older than (latest - retain).
func (db *DB) PruneOldBlocks(retain uint64) error {
	latest, err := db.GetLatestHeight()
	if err != nil || latest <= retain {
		return err // Nothing to prune
	}

	target := latest - retain

	// Iterate downwards to cleanly remove all traces of older blocks
	batch := db.conn.NewBatch()
	defer batch.Close()

	for i := target; i > 0; i-- {
		b, err := db.GetBlock(i)
		if err != nil {
			continue // Already pruned or missing
		}

		// Delete Header: b_<height>
		_ = batch.Delete(makeKey(PrefixBlocks, uint64ToBytes(i)), nil)

		// Delete Block Hash Index: h_<hash>
		_ = batch.Delete(makeKey(PrefixHashIdx, b.Hash), nil)

		// Delete Transaction Data & Indexes
		for _, hash := range b.TransactionHashes {
			_ = batch.Delete(makeKey(PrefixTxData, hash), nil)
			_ = batch.Delete(makeKey(PrefixTxIdx, hash), nil)
		}

		// Flush batch every 100 blocks to prevent memory explosion
		if i%100 == 0 {
			if err := batch.Commit(pebble.Sync); err != nil {
				return err
			}
			batch = db.conn.NewBatch() // Create a new batch for the next 100
		}
	}

	if err := batch.Commit(pebble.Sync); err != nil {
		return err
	}

	if db.bus != nil {
		_ = db.bus.Publish(event.Event{
			Type:    event.BlocksPruned,
			Source:  "Storage",
			Payload: target,
		})
	}
	return nil
}

//
// ==============================
// O(1) TX LOOKUP
// ==============================
//

func (db *DB) GetTransactionByHash(hash string) (*block.Block, int, error) {
	key := makeKey(PrefixTxIdx, []byte(hash))
	data, closer, err := db.conn.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, 0, errors.New("transaction not found")
		}
		return nil, 0, err
	}
	defer closer.Close()

	var entry struct {
		Height uint64
		Index  int
	}

	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, 0, err
	}

	blockObj, err := db.GetBlock(entry.Height)
	if err != nil {
		return nil, 0, err
	}

	return blockObj, entry.Index, nil
}
