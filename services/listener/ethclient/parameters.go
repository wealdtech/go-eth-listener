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

// Package ethclient is a listener that listens to an Ethereum client.
package ethclient

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/wealdtech/go-eth-listener/handlers"
	"github.com/wealdtech/go-eth-listener/services/metrics"
	nullmetrics "github.com/wealdtech/go-eth-listener/services/metrics/null"
)

type parameters struct {
	logLevel       zerolog.Level
	clientLogLevel zerolog.Level
	monitor        metrics.Service
	metadataDBPath string
	address        string
	timeout        time.Duration
	blockDelay     uint32
	blockSpecifier string
	earliestBlock  int32
	blockTriggers  []*handlers.BlockTrigger
	txTriggers     []*handlers.TxTrigger
	eventTriggers  []*handlers.EventTrigger
	interval       time.Duration
}

// Parameter is the interface for service parameters.
type Parameter interface {
	apply(p *parameters)
}

type parameterFunc func(*parameters)

func (f parameterFunc) apply(p *parameters) {
	f(p)
}

// WithLogLevel sets the log level for the listener.
func WithLogLevel(logLevel zerolog.Level) Parameter {
	return parameterFunc(func(p *parameters) {
		p.logLevel = logLevel
	})
}

// WithClientLogLevel sets the log level for the clients used by the listener.
func WithClientLogLevel(logLevel zerolog.Level) Parameter {
	return parameterFunc(func(p *parameters) {
		p.clientLogLevel = logLevel
	})
}

// WithMonitor sets the metrics monitor.
func WithMonitor(monitor metrics.Service) Parameter {
	return parameterFunc(func(p *parameters) {
		p.monitor = monitor
	})
}

// WithMetadataDBPath sets the path of the metadata database.
func WithMetadataDBPath(path string) Parameter {
	return parameterFunc(func(p *parameters) {
		p.metadataDBPath = path
	})
}

// WithAddress sets the address of the Ethereum client.
func WithAddress(address string) Parameter {
	return parameterFunc(func(p *parameters) {
		p.address = address
	})
}

// WithTimeout sets the timeout for requests made to the Ethereum client.
func WithTimeout(timeout time.Duration) Parameter {
	return parameterFunc(func(p *parameters) {
		p.timeout = timeout
	})
}

// WithBlockDelay sets the number of blocks to delay before
// passing on to the handlers, allowing avoidance of reorgs.
// Ignored if block specifier is provided.
func WithBlockDelay(delay uint32) Parameter {
	return parameterFunc(func(p *parameters) {
		p.blockDelay = delay
	})
}

// WithBlockSpecifier sets the specifier for the block to handle.
// This override block delay if supplied.
func WithBlockSpecifier(specifier string) Parameter {
	return parameterFunc(func(p *parameters) {
		p.blockSpecifier = specifier
	})
}

// WithEarliestBlock sets the block number from which to start listening.
func WithEarliestBlock(block int32) Parameter {
	return parameterFunc(func(p *parameters) {
		p.earliestBlock = block
	})
}

// WithBlockTriggers sets the block triggers for the listener.
func WithBlockTriggers(triggers []*handlers.BlockTrigger) Parameter {
	return parameterFunc(func(p *parameters) {
		p.blockTriggers = triggers
	})
}

// WithTxTriggers sets the transaction triggers for the listener.
func WithTxTriggers(triggers []*handlers.TxTrigger) Parameter {
	return parameterFunc(func(p *parameters) {
		p.txTriggers = triggers
	})
}

// WithEventTriggers sets the event triggers for the listener.
func WithEventTriggers(triggers []*handlers.EventTrigger) Parameter {
	return parameterFunc(func(p *parameters) {
		p.eventTriggers = triggers
	})
}

// WithInterval sets the interval between polls.
func WithInterval(interval time.Duration) Parameter {
	return parameterFunc(func(p *parameters) {
		p.interval = interval
	})
}

// parseAndCheckParameters parses and checks parameters to ensure that mandatory parameters are present and correct.
func parseAndCheckParameters(params ...Parameter) (*parameters, error) {
	parameters := parameters{
		logLevel:       zerolog.GlobalLevel(),
		clientLogLevel: zerolog.GlobalLevel(),
		monitor:        nullmetrics.New(),
		earliestBlock:  -1,
	}
	for _, p := range params {
		if p != nil {
			p.apply(&parameters)
		}
	}

	if parameters.monitor == nil {
		return nil, errors.New("no monitor specified")
	}
	if parameters.timeout == 0 {
		return nil, errors.New("no timeout specified")
	}
	if parameters.address == "" {
		return nil, errors.New("no address specified")
	}
	if parameters.metadataDBPath == "" {
		return nil, errors.New("no metadata db path specified")
	}
	if err := checkTriggerParameters(&parameters); err != nil {
		return nil, err
	}
	if parameters.interval == 0 {
		return nil, errors.New("no interval specified")
	}

	validBlockSpecifiers := map[string]struct{}{
		"":          {},
		"latest":    {},
		"safe":      {},
		"finalized": {},
	}
	if _, exists := validBlockSpecifiers[strings.ToLower(parameters.blockSpecifier)]; !exists {
		return nil, fmt.Errorf("unsupported block specifier %s", parameters.blockSpecifier)
	}

	return &parameters, nil
}

func checkTriggerParameters(parameters *parameters) error {
	for _, blockTrigger := range parameters.blockTriggers {
		if blockTrigger.Name == "" {
			return errors.New("no block trigger name specified")
		}
		if blockTrigger.Handler == nil {
			return errors.New("no block trigger handler specified")
		}
	}
	for _, txTrigger := range parameters.txTriggers {
		if txTrigger.Name == "" {
			return errors.New("no transaction trigger name specified")
		}
		if txTrigger.Handler == nil {
			return errors.New("no transaction trigger handler specified")
		}
	}
	for _, eventTrigger := range parameters.eventTriggers {
		if eventTrigger.Name == "" {
			return errors.New("no event trigger name specified")
		}
		if eventTrigger.Handler == nil {
			return errors.New("no event trigger handler specified")
		}
	}

	return nil
}
