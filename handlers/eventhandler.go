package handlers

import (
	"context"

	"github.com/attestantio/go-execution-client/spec"
	"github.com/attestantio/go-execution-client/types"
)

// EventTrigger is a trigger for an event.
type EventTrigger struct {
	Name string
	// Source is a static address to use for event addresses.
	Source *types.Address
	// SourceResolver is a dynamic resolver use for event addresses.
	SourceResolver SourceResolver
	Topics         []types.Hash
	EarliestBlock  uint32
	Handler        EventHandler
}

// SourceResolver defines the methods that need to be implemented to resolve sources.
type SourceResolver interface {
	// Resolve resolves a source for events.
	Resolve(ctx context.Context) (*types.Address, error)
}

// EventHandlerFunc defines the handler function.
type EventHandlerFunc func(ctx context.Context, event *spec.BerlinTransactionEvent, trigger *EventTrigger)

// EventHandler defines the methods that need to be implemented to handle events.
type EventHandler interface {
	HandleEvent(ctx context.Context, event *spec.BerlinTransactionEvent, trigger *EventTrigger)
}
