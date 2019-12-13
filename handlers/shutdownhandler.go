package handlers

import (
	"github.com/wealdtech/go-eth-listener/shared"
)

// ShutdownHandlerFunc defines the handler function
type ShutdownHandlerFunc func(*shared.AppContext)

// Handle handles a transaction
func (f ShutdownHandlerFunc) Handle(actx *shared.AppContext) {
	f(actx)
}

// ShutdownHandler defines the methods that need to be implemented to handle shutdown
type ShutdownHandler interface {
	Handle(*shared.AppContext)
}
