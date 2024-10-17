// Copyright © 2023 Weald Technology Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ethclient

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	execclient "github.com/attestantio/go-execution-client"
	jsonrpcexecclient "github.com/attestantio/go-execution-client/jsonrpc"
	"github.com/cockroachdb/pebble"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	"github.com/wealdtech/go-eth-listener/handlers"
)

// Service is a listener that listens to an Ethereum client.
type Service struct {
	log                 zerolog.Logger
	chainHeightProvider execclient.ChainHeightProvider
	blocksProvider      execclient.BlocksProvider
	eventsProvider      execclient.EventsProvider
	blockTriggers       []*handlers.BlockTrigger
	txTriggers          []*handlers.TxTrigger
	eventTriggers       []*handlers.EventTrigger
	interval            time.Duration
	blockDelay          uint32
	blockSpecifier      string
	earliestBlock       int32
	metadataDB          *pebble.DB
	metadataDBMu        sync.Mutex
	metadataDBOpen      atomic.Bool
}

// New creates a new service.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, err
	}

	// Set logging.
	log := zerologger.With().Str("service", "listener").Str("impl", "ethclient").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	if err := registerMetrics(ctx, parameters.monitor); err != nil {
		return nil, err
	}

	chainHeightProvider, blocksProvider, eventsProvider, err := setupProviders(ctx, parameters)
	if err != nil {
		return nil, err
	}

	metadataDB, err := pebble.Open(parameters.metadataDBPath, &pebble.Options{})
	if err != nil {
		return nil, errors.Join(errors.New("failed to start metadata database"), err)
	}

	s := &Service{
		log:                 log,
		metadataDB:          metadataDB,
		blocksProvider:      blocksProvider,
		eventsProvider:      eventsProvider,
		blockTriggers:       parameters.blockTriggers,
		txTriggers:          parameters.txTriggers,
		eventTriggers:       parameters.eventTriggers,
		blockDelay:          parameters.blockDelay,
		blockSpecifier:      parameters.blockSpecifier,
		earliestBlock:       parameters.earliestBlock,
		chainHeightProvider: chainHeightProvider,
		interval:            parameters.interval,
	}

	// Note that the metadata DB is open.
	s.metadataDBOpen.Store(true)

	// Close the database on context done.
	go func(ctx context.Context, metadataDB *pebble.DB) {
		<-ctx.Done()
		s.metadataDBMu.Lock()
		err := metadataDB.Close()
		s.metadataDBOpen.Store(false)
		s.metadataDBMu.Unlock()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to close pebble")
		}
	}(ctx, metadataDB)

	// Kick off the listener.
	go s.listener(ctx)

	return s, nil
}

func setupProviders(ctx context.Context,
	parameters *parameters,
) (
	execclient.ChainHeightProvider,
	execclient.BlocksProvider,
	execclient.EventsProvider,
	error,
) {
	client, err := jsonrpcexecclient.New(ctx,
		jsonrpcexecclient.WithLogLevel(parameters.clientLogLevel),
		jsonrpcexecclient.WithAddress(parameters.address),
		jsonrpcexecclient.WithTimeout(parameters.timeout),
	)
	if err != nil {
		return nil, nil, nil, errors.Join(errors.New("failed to connect to Ethereum client"), err)
	}
	chainHeightProvider, isProvider := client.(execclient.ChainHeightProvider)
	if !isProvider {
		return nil, nil, nil, errors.New("client does not provide chain height")
	}
	blocksProvider, isProvider := client.(execclient.BlocksProvider)
	if !isProvider {
		return nil, nil, nil, errors.New("client does not provide blocks")
	}
	eventsProvider, isProvider := client.(execclient.EventsProvider)
	if !isProvider {
		return nil, nil, nil, errors.New("client does not provide events")
	}

	return chainHeightProvider, blocksProvider, eventsProvider, nil
}
