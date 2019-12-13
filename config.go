package listener

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/wealdtech/go-eth-listener/handlers"
)

// Config is the configuration of the handlers
type Config struct {
	// Connection is a connection to an Ethereum backend
	Connection *ethclient.Client
	// From is the block from which to start listening, if undefined.
	// nil means start from the latest block
	From *big.Int
	// Delay is the number of blocks to delay (avoids reorganisations)
	Delay uint
	// Timeout is the time after which attempts to obtain data will fail
	Timeout time.Duration
	// PollInterval is the interval between polling tasks
	PollInterval time.Duration
	// Extra is extra configuration supplied by the calling code
	Extra interface{}
	// InitHandlers are handlers fired when the listener starts
	InitHandlers handlers.InitHandler
	// EventHandlers are handlers fired when new events are received
	EventHandlers handlers.EventHandler
	// BlkHandlers are handlers fired when new blocks are received
	BlkHandlers handlers.BlkHandler
	// TxHandlers are handlers fired when new transactions are received as part of blocks
	TxHandlers handlers.TxHandler
	// PendingTxHandlers are handlers fired when new transactions are received in to the transaction pool
	PendingTxHandlers handlers.TxHandler
	// PollHandlers are handlers fired periodically
	PollHandlers handlers.PollHandler
	// ShutdownHandlers are handlers fired when the listener stops
	ShutdownHandlers handlers.ShutdownHandler
}
