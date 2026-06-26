package event

import (
	"fmt"
	"sync"
)

// EventBus implements the AES Internal Domain Event Bus.
// It is synchronous, thread-safe, and preserves registration order dispatching.
type EventBus struct {
	mu          sync.RWMutex
	subscribers []Subscriber
}

// NewEventBus creates a new synchronous in-memory event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make([]Subscriber, 0),
	}
}

// Subscribe registers a new subscriber. Order is preserved natively by slice insertion order.
func (b *EventBus) Subscribe(sub Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = append(b.subscribers, sub)
}

// Publish synchronously dispatches the event to all interested subscribers in registration order.
// Deterministic ordering is guaranteed by the slice structure and synchronous blocking execution.
func (b *EventBus) Publish(e Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		if isInterested(sub, e.Type) {
			if err := sub.Handle(e); err != nil {
				return fmt.Errorf("subscriber %s failed handling event %s: %w", sub.Name(), e.Type, err)
			}
		}
	}
	return nil
}

// isInterested is a fast helper to check if a subscriber wants a specific event type
func isInterested(sub Subscriber, eType EventType) bool {
	for _, t := range sub.InterestedIn() {
		if t == eType {
			return true
		}
	}
	return false
}
