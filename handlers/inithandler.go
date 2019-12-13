package handlers

import (
	"github.com/wealdtech/go-eth-listener/shared"
)

// InitHandlerFunc defines the handler function
type InitHandlerFunc func(*shared.AppContext)

// Handle handles a transaction
func (f InitHandlerFunc) Handle(actx *shared.AppContext) {
	f(actx)
}

// InitHandler defines the methods that need to be implemented to handle initialisation
type InitHandler interface {
	Handle(*shared.AppContext)
}
