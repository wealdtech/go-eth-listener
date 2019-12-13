package handlers

import (
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/go-eth-listener/shared"
)

// LogInit is called when the log module is initialised.
func LogInit(h InitHandler) InitHandler {
	return InitHandlerFunc(func(actx *shared.AppContext) {
		log.WithFields(log.Fields{"module": "log"}).Info("Initialised")
		if h != nil {
			h.Handle(actx)
		}
	})
}

// LogShutdown is called when the log module is shut down.
func LogShutdown(h ShutdownHandler) ShutdownHandler {
	return ShutdownHandlerFunc(func(actx *shared.AppContext) {
		log.WithFields(log.Fields{"module": "log"}).Info("Shutdown")
		if h != nil {
			h.Handle(actx)
		}
	})
}

// LogPendingTx is called when the log module receives a pending transaction.
func LogPendingTx(h TxHandler) TxHandler {
	return TxHandlerFunc(func(actx *shared.AppContext, tx *types.Transaction) {
		log.WithFields(log.Fields{"transaction": tx.Hash()}).Info("Pending transaction received")
		if h != nil {
			h.Handle(actx, tx)
		}
	})
}

// LogTx is called when the log module receives a transaction.
func LogTx(h TxHandler) TxHandler {
	return TxHandlerFunc(func(actx *shared.AppContext, tx *types.Transaction) {
		log.WithFields(log.Fields{"hash": tx.Hash()}).Info("Transaction received")
		if h != nil {
			h.Handle(actx, tx)
		}
	})
}

// LogBlk is called when the log module receives a block.
func LogBlk(h BlkHandler) BlkHandler {
	return BlkHandlerFunc(func(actx *shared.AppContext, blk *types.Block) {
		log.WithFields(log.Fields{"hash": blk.Hash(), "number": blk.Number()}).Info("Block received")
		if h != nil {
			h.Handle(actx, blk)
		}
	})
}

// LogPoll is called when the log module receives a poll.
func LogPoll(h PollHandler) PollHandler {
	return PollHandlerFunc(func(actx *shared.AppContext, tick uint64) {
		log.WithFields(log.Fields{"tick": tick}).Info("Tick")
		if h != nil {
			h.Handle(actx, tick)
		}
	})
}

// LogEvent is called when the log module receives an event.
func LogEvent(h EventHandler) EventHandler {
	return EventHandlerFunc(func(actx *shared.AppContext, event *types.Log) {
		log.WithFields(log.Fields{"event": event}).Info("Event received")
		if h != nil {
			h.Handle(actx, event)
		}
	})
}
