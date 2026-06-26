package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/sureshKrishna05/aegisq-framework/core/event"
)

// AdaptiveCompactionController triggers PebbleDB compactions during idle blockchain periods
// rather than competing for CPU during active consensus windows.
type AdaptiveCompactionController struct {
	db     Store
	cancel context.CancelFunc
}

func NewAdaptiveCompactionController(db Store) *AdaptiveCompactionController {
	return &AdaptiveCompactionController{db: db}
}

// Start listens to the event bus and triggers compactions when conditions are right
func (c *AdaptiveCompactionController) Start(bus *event.EventBus) {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	go func() {
		// Run a naive timer for now to simulate idle-window compaction
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Println("[Compaction Controller] Triggering adaptive background compaction")
				_ = c.db.CompactDB()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop gracefully halts the adaptive compaction routine
func (c *AdaptiveCompactionController) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}
