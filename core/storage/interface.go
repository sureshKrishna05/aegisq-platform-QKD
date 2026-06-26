package storage

import (
	"github.com/sureshKrishna05/aegisq-framework/core/block"
)

// Store defines the formal Database Abstraction Layer (ABS) interface
// allowing underlying storage engines to be hot-swapped (e.g. PebbleDB, LevelDB).
type Store interface {
	// Lifecycle
	Close() error
	CheckIntegrity() error

	// Blocks
	SaveBlock(b *block.Block) error
	GetBlock(height uint64) (*block.Block, error)
	GetLatestHeight() (uint64, error)

	// O(1) Lookups
	GetTransactionByHash(hash string) (*block.Block, int, error)

	// Phase 4: Lifecycle & Maintenance
	CreateSnapshot(destDir string) error
	PruneOldBlocks(retain uint64) error
	CompactDB() error
}
