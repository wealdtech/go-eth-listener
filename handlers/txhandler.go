package handlers

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-eth-listener/shared"
)

// TxHandlerFunc defines the handler function
type TxHandlerFunc func(*shared.AppContext, *types.Transaction)

// Handle handles a transaction
func (f TxHandlerFunc) Handle(actx *shared.AppContext, tx *types.Transaction) {
	f(actx, tx)
}

// TxHandler defines the methods that need to be implemented to handle transactions
type TxHandler interface {
	Handle(*shared.AppContext, *types.Transaction)
}
