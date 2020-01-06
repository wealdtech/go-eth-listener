package handlers

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-eth-listener/shared"
)

// TxHandlerFunc defines the handler function
type TxHandlerFunc func(*shared.AppContext, *types.Block, *types.Transaction)

// Handle handles a transaction
func (f TxHandlerFunc) Handle(actx *shared.AppContext, blk *types.Block, tx *types.Transaction) {
	f(actx, blk, tx)
}

// TxHandler defines the methods that need to be implemented to handle transactions
type TxHandler interface {
	Handle(*shared.AppContext, *types.Block, *types.Transaction)
}
