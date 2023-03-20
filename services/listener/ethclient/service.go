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
	"time"

	execclient "github.com/attestantio/go-execution-client"
	jsonrpcexecclient "github.com/attestantio/go-execution-client/jsonrpc"
	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	"github.com/wealdtech/go-eth-listener/handlers"
	"github.com/wealdtech/go-eth-listener/util"
)

// Service is a listener that listens to an Ethereum client.
type Service struct {
	log                 zerolog.Logger
	chainHeightProvider execclient.ChainHeightProvider
	eventsProvider      execclient.EventsProvider
	blockTriggers       []*handlers.BlockTrigger
	txTriggers          []*handlers.TxTrigger
	eventTriggers       []*handlers.EventTrigger
	interval            time.Duration
	blockDelay          uint32
	metadataDB          *pebble.DB
}

// New creates a new service.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log := zerologger.With().Str("service", "listener").Str("impl", "ethclient").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	client, err := jsonrpcexecclient.New(ctx,
		jsonrpcexecclient.WithLogLevel(util.LogLevel("execclient")),
		jsonrpcexecclient.WithAddress(parameters.address),
		jsonrpcexecclient.WithTimeout(parameters.timeout),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Ethereum client")
	}
	chainHeightProvider, isProvider := client.(execclient.ChainHeightProvider)
	if !isProvider {
		return nil, errors.New("client does not provide chain height")
	}
	eventsProvider, isProvider := client.(execclient.EventsProvider)
	if !isProvider {
		return nil, errors.New("client does not provide events")
	}

	s := &Service{
		log:                 log,
		eventsProvider:      eventsProvider,
		blockTriggers:       parameters.blockTriggers,
		txTriggers:          parameters.txTriggers,
		eventTriggers:       parameters.eventTriggers,
		blockDelay:          parameters.blockDelay,
		metadataDB:          parameters.metadataDB,
		chainHeightProvider: chainHeightProvider,
		interval:            parameters.interval,
	}

	// Kick off the listener.
	go s.listener(ctx)

	return s, nil
}
