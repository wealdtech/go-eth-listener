package handlers

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-eth-listener/shared"
)

// EventHandlerFunc defines the handler function
type EventHandlerFunc func(*shared.AppContext, *types.Log)

// Handle handles a Block
func (f EventHandlerFunc) Handle(actx *shared.AppContext, event *types.Log) {
	f(actx, event)
}

// EventHandler defines the methods that need to be implemented to handle events
type EventHandler interface {
	Handle(*shared.AppContext, *types.Log)
}
