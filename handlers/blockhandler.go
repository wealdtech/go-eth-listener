package handlers

import (
	"context"

	"github.com/attestantio/go-execution-client/spec"
)

// BlockTrigger is a trigger for a block.
type BlockTrigger struct {
	Name    string
	Handler BlockHandler
}

// BlockHandlerFunc defines the handler function.
type BlockHandlerFunc func(ctx context.Context, block *spec.Block, trigger *BlockTrigger)

// BlockHandler defines the methods that need to be implemented to handle events.
type BlockHandler interface {
	HandleBlock(ctx context.Context, block *spec.Block, trigger *BlockTrigger)
}
