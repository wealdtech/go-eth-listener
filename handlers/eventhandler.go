package handlers

import (
	"github.com/attestantio/go-execution-client/spec"
	"github.com/attestantio/go-execution-client/types"
)

// EventTrigger is a trigger for an event.
type EventTrigger struct {
	Name          string
	Source        *types.Address
	Topics        []types.Hash
	EarliestBlock uint32
	Handler       EventHandler
}

// EventHandlerFunc defines the handler function.
type EventHandlerFunc func(event *spec.BerlinTransactionEvent, trigger *EventTrigger)

// EventHandler defines the methods that need to be implemented to handle events.
type EventHandler interface {
	HandleEvent(event *spec.BerlinTransactionEvent, trigger *EventTrigger)
}
