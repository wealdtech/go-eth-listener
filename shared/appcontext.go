package shared

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// AppContext is a structure holding connections to external entities
type AppContext struct {
	// Connection is a connection to an Ethereum node
	Connection *ethclient.Client
	// Timeout is the time after which attempts to obtain data will fail
	Timeout time.Duration
	// ChainID is the ID of the Ethereum chain to which we are connected
	ChainID *big.Int
	// Extra is extra configuration supplied by the calling code
	Extra interface{}
}
