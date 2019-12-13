package handlers

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-eth-listener/shared"
)

// BlkHandlerFunc defines the handler function
type BlkHandlerFunc func(*shared.AppContext, *types.Block)

// Handle handles a Block
func (f BlkHandlerFunc) Handle(actx *shared.AppContext, blk *types.Block) {
	f(actx, blk)
}

// BlkHandler defines the methods that need to be implemented to handle blocks
type BlkHandler interface {
	Handle(*shared.AppContext, *types.Block)
}
