package event

import (
	"time"
)

// EventType categorizes internal domain events
type EventType string

// Defined Event Categories & Types according to AES Specification
const (
	// Node Events
	NodeStarted          EventType = "NodeStarted"
	RecoveryCompleted    EventType = "RecoveryCompleted"
	IntegrityCheckPassed EventType = "IntegrityCheckPassed"

	// Transaction Events
	TransactionReceived  EventType = "TransactionReceived"
	TransactionValidated EventType = "TransactionValidated"
	TransactionRejected  EventType = "TransactionRejected"

	// Consensus Events
	LeaderSelected EventType = "LeaderSelected"
	BlockProposed  EventType = "BlockProposed"
	PrepareReached EventType = "PrepareReached"
	CommitReached  EventType = "CommitReached"
	BlockCommitted EventType = "BlockCommitted"

	// Ledger Events
	BlockPersisted  EventType = "BlockPersisted"
	SnapshotCreated EventType = "SnapshotCreated"
	BlocksPruned    EventType = "BlocksPruned"
)

// Event structure representing an internal domain event
type Event struct {
	ID            string
	Type          EventType
	Timestamp     time.Time
	Source        string
	CorrelationID string
	Payload       interface{}
}

// Subscriber defines the contract for any component listening to the event bus
type Subscriber interface {
	Name() string
	InterestedIn() []EventType
	Handle(e Event) error
}
