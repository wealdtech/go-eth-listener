package listener

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/go-eth-listener/shared"
)

// Listen listens to a blockchain and triggers functions as new blocks, transactions etc. arrive
func Listen(config *Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	var err error
	chainID, err = config.Connection.NetworkID(ctx)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to obtain chain ID")
		return err
	}

	actx := &shared.AppContext{
		Ctx:        ctx,
		Connection: config.Connection,
		Timeout:    config.Timeout,
		ChainID:    chainID,
		Extra:      config.Extra,
	}

	initProcessor(config)
	firstRun := initCheckpoint(actx)
	log.WithFields(log.Fields{"firstrun": firstRun}).Info("First run check")

	// Initialisation handlers
	if config.InitHandlers != nil {
		config.InitHandlers.Handle(actx)
	}

	if !firstRun || config.From != nil {
		// Catch up on missed blocks
		curBlock := new(big.Int)
		if config.From != nil {
			curBlock.Set(config.From)
		} else {
			curBlock.Set(checkpointBlock)
			if !firstRun {
				curBlock.Add(curBlock, big.NewInt(1))
			}
		}
		log.WithFields(log.Fields{"from": curBlock}).Info("Catching up on old blocks")

		for ; ; curBlock.Add(curBlock, big.NewInt(1)) {
			blk, err := config.Connection.BlockByNumber(ctx, curBlock)
			if err != nil {
				// TODO Assuming we have hit the current block; should confirm this is the case
				break
			}
			processBlock(actx, config, blk)
		}
	}

	// Catch new blocks
	blkHdrCh := make(chan *types.Header)
	if config.BlkHandlers != nil || config.TxHandlers != nil || config.EventHandlers != nil {
		if config.BlkHandlers != nil {
			_, err := config.Connection.SubscribeNewHead(ctx, blkHdrCh)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Fatal("failed to subscribe to block updates")
			}
		}
	}

	// Set up polling
	if config.PollHandlers != nil {
		go poll(actx, config)
	}

	// Catch pending transactions
	pendingTxCh := make(chan *types.Transaction)
	if config.PendingTxHandlers != nil {
		log.Warn("pending transactions not implemented")
		//		pendingTxCtx, pendingTxCancel := context.WithTimeout(context.Background(), config.Timeout)
		//		defer pendingTxCancel()
		//		_, err := config.Connection.SubscribePendingTransactions(pendingTxCtx, pendingTxCh)
		//		if err != nil {
		//			log.WithFields(log.Fields{"error": err}).Fatal("failed to subscribe to pending transactions")
		//		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Loop
	for {
		select {
		case pendingTx := <-pendingTxCh:
			config.PendingTxHandlers.Handle(actx, pendingTx)
		case blkHdr := <-blkHdrCh:
			// Obtain block from the block header
			blk, err := config.Connection.BlockByNumber(ctx, blkHdr.Number)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to obtain block")
				continue
			}
			processBlock(actx, config, blk)
		case <-interrupt:
			log.Info("Shutdown")
			if config.ShutdownHandlers != nil {
				config.ShutdownHandlers.Handle(actx)
			}
			os.Exit(0)
		}
	}
}

func poll(actx *shared.AppContext, config *Config) {
	tick := uint64(0)
	ticker := time.NewTicker(config.PollInterval)
	for {
		select {
		case <-ticker.C:
			config.PollHandlers.Handle(actx, tick)
			tick++
		case <-actx.Ctx.Done():
			ticker.Stop()
			return
		}
	}
}
