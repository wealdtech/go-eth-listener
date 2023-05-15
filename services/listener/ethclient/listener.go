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
	"bytes"
	"context"
	"time"

	"github.com/attestantio/go-execution-client/api"
	"github.com/attestantio/go-execution-client/types"
	executil "github.com/attestantio/go-execution-client/util"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Maximum number of blocks to fetch for events.
var maxBlocksForEvents = uint32(100)

func (s *Service) listener(ctx context.Context,
) {
	// Start with a poll.
	s.poll(ctx)

	// Now loop until context is cancelled.
	for {
		select {
		case <-time.After(s.interval):
			s.poll(ctx)
		case <-ctx.Done():
			s.log.Debug().Msg("Context cancelled")
			return
		}
	}
}

func (s *Service) poll(ctx context.Context) {
	var to uint32
	// Select the highest block to work with.
	if s.blockSpecifier != "" {
		log.Trace().Str("specifier", s.blockSpecifier).Msg("Fetching chain height for specifier")
		block, err := s.blocksProvider.Block(ctx, s.blockSpecifier)
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to obtain block")
			monitorFailure()
			return
		}
		to = block.Number()
	} else {
		chainHeight, err := s.chainHeightProvider.ChainHeight(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get chain height for event poll")
			monitorFailure()
			return
		}
		to = chainHeight - s.blockDelay
	}
	s.log.Trace().Uint32("to", to).Msg("Selected highest block")

	switch {
	case len(s.blockTriggers) > 0:
		// We have block triggers, fetch full blocks.
		if err := s.pollBlocks(ctx, to); err != nil {
			log.Error().Err(err).Msg("Failure to poll blocks")
			monitorFailure()
		}
	case len(s.txTriggers) > 0:
		// We have transaction triggers, fetch full blocks.
		if err := s.pollTxs(ctx, to); err != nil {
			log.Error().Err(err).Msg("Failure to poll transactions")
			monitorFailure()
		}
	case len(s.eventTriggers) > 0:
		// We have event triggers, fetch events only.
		if err := s.pollEvents(ctx, to); err != nil {
			log.Error().Err(err).Msg("Failure to poll events")
			monitorFailure()
		}
	}

	monitorLatestBlock(to)
}

func (s *Service) pollBlocks(ctx context.Context,
	to uint32,
) error {
	md, err := s.getBlocksMetadata(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get metadata for block poll")
	}

	from := uint32(md.LatestBlock + 1)

	if from > to {
		s.log.Trace().Uint32("from", from).Uint32("to", to).Msg("Not fetching blocks")
		return nil
	}

	for height := from; height <= to; height++ {
		block, err := s.blocksProvider.Block(ctx, executil.MarshalUint32(height))
		if err != nil {
			return errors.Wrap(err, "failed to obtain block")
		}
		for _, trigger := range s.blockTriggers {
			trigger.Handler.HandleBlock(ctx, block, trigger)
		}

		md.LatestBlock = int32(height)
		if err := s.setBlocksMetadata(ctx, md); err != nil {
			return errors.Wrap(err, "failed to set metadata after block poll")
		}
	}

	return nil
}

func (s *Service) pollTxs(ctx context.Context,
	to uint32,
) error {
	md, err := s.getTransactionsMetadata(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get metadata for transaction poll")
	}

	from := uint32(md.LatestBlock + 1)

	if from > to {
		s.log.Trace().Uint32("from", from).Uint32("to", to).Msg("Not fetching blocks for transactions")
		return nil
	}

	for height := from; height <= to; height++ {
		block, err := s.blocksProvider.Block(ctx, executil.MarshalUint32(height))
		if err != nil {
			return errors.Wrap(err, "failed to obtain block for transactions")
		}
		for _, trigger := range s.txTriggers {
			if block.Number() < trigger.EarliestBlock {
				log.Trace().Str("trigger", trigger.Name).Uint32("block_height", block.Number()).Msg("Block too early; ignoring")
				continue
			}
			for i, tx := range block.Transactions() {
				if trigger.From != nil {
					txFrom := tx.From()
					if !bytes.Equal(trigger.From[:], txFrom[:]) {
						log.Trace().Str("trigger", trigger.Name).Uint32("block_height", block.Number()).Int("index", i).Msg("From does not match; ignoring")
						continue
					}
				}
				if trigger.To != nil {
					txTo := tx.To()
					if !bytes.Equal(trigger.To[:], txTo[:]) {
						log.Trace().Str("trigger", trigger.Name).Uint32("block_height", block.Number()).Int("index", i).Msg("To does not match; ignoring")
						continue
					}
				}
				trigger.Handler.HandleTx(ctx, tx, trigger)
			}
		}

		md.LatestBlock = int32(height)
		if err := s.setTransactionsMetadata(ctx, md); err != nil {
			return errors.Wrap(err, "failed to set metadata after trasaction poll")
		}
	}

	return nil
}

func (s *Service) pollEvents(ctx context.Context,
	to uint32,
) error {
	md, err := s.getEventsMetadata(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get metadata for event poll")
	}

	// Need to run each trigger separately.
	for _, trigger := range s.eventTriggers {
		// Obtain the last block we examined for this trigger, or
		// use the earliest block as defined in the trigger.
		from := trigger.EarliestBlock
		if _, exists := md.LatestBlocks[trigger.Name]; exists {
			next := md.LatestBlocks[trigger.Name] + 1
			if next > from {
				from = next
			}
		}
		if from > to {
			s.log.Trace().Str("trigger", trigger.Name).Uint32("from", from).Uint32("to", to).Msg("Not fetching events")
			return nil
		}

		// Resolve the source.
		var source *types.Address
		switch {
		case trigger.SourceResolver != nil:
			source, err = trigger.SourceResolver.Resolve(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to resolve source")
			}
		case trigger.Source != nil:
			source = trigger.Source
		default:
			return errors.New("no source")
		}
		if source == nil {
			return errors.New("source resolution returned nil")
		}

		if to+1-from > maxBlocksForEvents {
			to = from + maxBlocksForEvents - 1
		}
		s.log.Trace().Stringer("source", source).Str("trigger", trigger.Name).Uint32("from", from).Uint32("to", to).Msg("Fetching events")

		filter := &api.EventsFilter{
			FromBlock: executil.MarshalUint32(from),
			ToBlock:   executil.MarshalUint32(to),
			Address:   source,
			Topics:    trigger.Topics,
		}
		events, err := s.eventsProvider.Events(ctx, filter)
		if err != nil {
			return errors.Wrap(err, "failed to obtain events")
		}
		for _, event := range events {
			trigger.Handler.HandleEvent(ctx, event, trigger)
		}

		md.LatestBlocks[trigger.Name] = to
		if err := s.setEventsMetadata(ctx, md); err != nil {
			return errors.Wrap(err, "failed to set metadata after event poll")
		}
	}

	return nil
}
