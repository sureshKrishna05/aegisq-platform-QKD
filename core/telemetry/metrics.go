package telemetry

import (
	"sync/atomic"
)

// Metric registers for Performance Engineering
var (
	TotalBlocksPersisted uint64
	TotalCacheHits       uint64
	TotalTransactions    uint64
)

// IncBlocks increments the block persist counter safely
func IncBlocks() {
	atomic.AddUint64(&TotalBlocksPersisted, 1)
}

// IncCacheHits increments the RAM cache hit counter safely
func IncCacheHits() {
	atomic.AddUint64(&TotalCacheHits, 1)
}

// IncTransactions increments the global transaction counter
func IncTransactions(count uint64) {
	atomic.AddUint64(&TotalTransactions, count)
}
