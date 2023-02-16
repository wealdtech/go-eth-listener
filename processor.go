package listener

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/go-eth-listener/shared"
)

var queue []*types.Block

func initProcessor(config *Config) {
	queue = make([]*types.Block, 0)
}

// processBlock processes a block, triggering handlers for each transaction and
// event in each transaction, along with handlers for the block itself.
//
// Ordering is important.  For each transaction the transaction is handled
// first, followed by any events generated by the transaction, followed by the
// block itself.
func processBlock(actx *shared.AppContext, config *Config, blk *types.Block) {
	if uint(len(queue)) < config.Delay {
		// Queue not full; maybe add this block and return
		if len(queue) == 0 || queue[len(queue)-1].NumberU64() < blk.NumberU64() {
			queue = append(queue, blk)
		}
		return
	}
	// Ensure this block is higher than the current last
	if config.Delay > 0 && queue[len(queue)-1].NumberU64() >= blk.NumberU64() {
		return
	}

	// Pull the earliest block
	var block *types.Block
	if config.Delay == 0 {
		block = blk
	} else {
		block = queue[0]
		queue = queue[1:]
		queue = append(queue, blk)
	}

	// Refetch the block in case it was overridden by a reorg
	// (only if we have a delay; otherwise all blocks are processed as they arrive)
	if config.Delay != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		defer cancel()
		var err error
		block, err = config.Connection.BlockByNumber(ctx, block.Number())
		if err != nil {
			log.WithError(err).Error("Failed to obtain block")
			return
		}
	}

	// Process the block's transactions
	if block.Transactions().Len() > 0 &&
		(config.TxHandlers != nil || config.EventHandlers != nil) {
		for _, tx := range block.Transactions() {
			if config.TxHandlers != nil {
				config.TxHandlers.Handle(actx, block, tx)
			}
			if config.EventHandlers != nil {
				ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
				defer cancel()
				receipt, err := config.Connection.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					log.WithError(err).Error("Failed to obtain block")
					continue
				} else {
					for _, log := range receipt.Logs {
						config.EventHandlers.Handle(actx, block, tx, log)
					}
				}
			}
		}
	}
	if config.BlkHandlers != nil {
		config.BlkHandlers.Handle(actx, block)
	}
	err := writeCheckpoint(actx.ChainID, block.Number())
	if err != nil {
		log.WithError(err).Error("Failed to write checkpoint")
	}
}