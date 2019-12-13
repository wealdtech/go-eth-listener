package handlers

import (
	"github.com/wealdtech/go-eth-listener/shared"
)

// PollHandlerFunc defines the handler function
type PollHandlerFunc func(*shared.AppContext, uint64)

// Handle handles a transaction
func (f PollHandlerFunc) Handle(actx *shared.AppContext, tick uint64) {
	f(actx, tick)
}

// PollHandler defines the methods that need to be implemented to handle ticks
type PollHandler interface {
	Handle(*shared.AppContext, uint64)
}
