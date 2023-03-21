package handlers

import (
	"context"

	"github.com/attestantio/go-execution-client/spec"
	"github.com/attestantio/go-execution-client/types"
)

// TxTrigger is a trigger for a transaction.
type TxTrigger struct {
	Name          string
	From          *types.Address
	To            *types.Address
	EarliestBlock uint32
	Handler       TxHandler
}

// TxHandlerFunc defines the handler function.
type TxHandlerFunc func(ctx context.Context, tx *spec.Transaction, trigger *TxTrigger)

// TxHandler defines the methods that need to be implemented to handle transactions.
type TxHandler interface {
	HandleTx(ctx context.Context, tx *spec.Transaction, trigger *TxTrigger)
}
