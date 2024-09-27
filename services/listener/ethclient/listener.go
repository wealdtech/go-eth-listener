// Copyright © 2023, 2024 Weald Technology Limited.
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
	"github.com/wealdtech/go-eth-listener/handlers"
)

// Maximum number of blocks to fetch for events.
const maxBlocksForEvents = uint32(100)

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
			s.log.Debug().Msg("Context done")
			return
		}
	}
}

func (s *Service) selectHighestBlock(ctx context.Context) (uint32, error) {
	var to uint32
	// Select the highest block with which to work, based on the specifier or the block delay.
	if s.blockSpecifier != "" {
		block, err := s.blocksProvider.Block(ctx, s.blockSpecifier)
		if err != nil {
			return 0, errors.Wrap(err, "failed to obtain block")
		}
		to = block.Number()
		s.log.Trace().Str("specifier", s.blockSpecifier).Uint32("height", to).Msg("Obtained chain height with specifier")
	} else {
		chainHeight, err := s.chainHeightProvider.ChainHeight(ctx)
		if err != nil {
			return 0, errors.Wrap(err, "failed to get chain height for event poll")
		}
		to = chainHeight - s.blockDelay
		s.log.Trace().Uint32("block_delay", s.blockDelay).Uint32("height", to).Msg("Obtained chain height with delay")
	}

	return to, nil
}

func (s *Service) poll(ctx context.Context) {
	to, err := s.selectHighestBlock(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to select highest block")
		monitorFailure()
		return
	}

	s.log.Trace().Uint32("to", to).Msg("Selected highest block")

	switch {
	case len(s.blockTriggers) > 0:
		// We have block triggers, fetch full blocks.
		s.log.Trace().Msg("Polling blocks")
		err = s.pollBlocks(ctx, to)
	case len(s.txTriggers) > 0:
		// We have transaction triggers, fetch full blocks.
		s.log.Trace().Msg("Polling blocks for transactions")
		err = s.pollTxs(ctx, to)
	case len(s.eventTriggers) > 0:
		// We have event triggers, fetch events only.
		s.log.Trace().Msg("Polling events")
		err = s.pollEvents(ctx, to)
	}
	if err != nil && ctx.Err() == nil {
		s.log.Error().Err(err).Msg("Poll failed")
		monitorFailure()
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

	from := s.calculateBlocksFrom(ctx, md)
	s.log.Trace().Uint32("from", from).Uint32("to", to).Msg("Polling blocks in range")
	if from > to {
		return nil
	}

	failed := make(map[string]bool)
	for height := from; height <= to; height++ {
		s.log.Trace().Uint32("block", height).Msg("Handling block")
		block, err := s.blocksProvider.Block(ctx, executil.MarshalUint32(height))
		if err != nil {
			return errors.Wrap(err, "failed to obtain block")
		}
		for _, trigger := range s.blockTriggers {
			if failed[trigger.Name] {
				// The trigger already reported a failure in this run, so don't run for future blocks.
				continue
			}
			if md.LatestBlocks[trigger.Name] >= int32(height) {
				// The trigger has already successfully processed this block.
				continue
			}
			if err := trigger.Handler.HandleBlock(ctx, block, trigger); err != nil {
				s.log.Debug().Str("trigger", trigger.Name).Uint32("block", height).Err(err).Msg("Trigger failed to handle block")
				// The trigger has reported a failure.  We stop here for this trigger and don't update its metadata.
				failed[trigger.Name] = true

				continue
			}
			md.LatestBlocks[trigger.Name] = int32(height)
		}

		if err := s.setBlocksMetadata(ctx, md); err != nil {
			return errors.Wrap(err, "failed to set metadata after block poll")
		}
	}

	return nil
}

const maxUint32 = uint32(0xffffffff)

// calculateBlocksFrom calculates the earliest block which we need to fetch.
func (s *Service) calculateBlocksFrom(_ context.Context, md *blocksMetadata) uint32 {
	var from uint32

	switch {
	case s.earliestBlock > -1:
		// There is a hard-coded earliest block passed to us in configuration, so we must start there.
		// We have to reset the metadata, otherwise blocks won't be reprocessed.
		from = uint32(s.earliestBlock)
		for name := range md.LatestBlocks {
			md.LatestBlocks[name] = s.earliestBlock - 1
		}
		s.earliestBlock = -1
	case len(md.LatestBlocks) > 0:
		// Work out the earliest block from our existing metadata.
		from = maxUint32
		for _, latest := range md.LatestBlocks {
			if from > uint32(latest+1) {
				from = uint32(latest + 1)
			}
		}
	default:
		// Means that there is no metadata or hard-coded block, so start from the beginning.
		from = 0
	}

	return from
}

func (s *Service) pollTxs(ctx context.Context,
	to uint32,
) error {
	md, err := s.getTransactionsMetadata(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get metadata for transaction poll")
	}

	from := uint32(md.LatestBlock + 1)
	if s.earliestBlock != -1 {
		from = uint32(s.earliestBlock)
		s.earliestBlock = -1
	}

	if from > to {
		s.log.Trace().Uint32("from", from).Uint32("to", to).Msg("Not fetching blocks for transactions")
		return nil
	}

	for height := from; height <= to; height++ {
		if err := s.pollBlockTxs(ctx, height); err != nil {
			return err
		}

		md.LatestBlock = int32(height)
		if err := s.setTransactionsMetadata(ctx, md); err != nil {
			return errors.Wrap(err, "failed to set metadata after trasaction poll")
		}
	}

	return nil
}

func (s *Service) pollBlockTxs(ctx context.Context, height uint32) error {
	block, err := s.blocksProvider.Block(ctx, executil.MarshalUint32(height))
	if err != nil {
		return errors.Wrap(err, "failed to obtain block for transactions")
	}
	log := s.log.With().Uint32("block_height", block.Number()).Logger()
	for _, trigger := range s.txTriggers {
		log := log.With().Str("trigger", trigger.Name).Logger()
		if block.Number() < trigger.EarliestBlock {
			log.Trace().Msg("Block too early; ignoring")
			continue
		}
		for i, tx := range block.Transactions() {
			if trigger.From != nil {
				txFrom := tx.From()
				if !bytes.Equal(trigger.From[:], txFrom[:]) {
					log.Trace().Int("index", i).Msg("From does not match; ignoring")
					continue
				}
			}
			if trigger.To != nil {
				txTo := tx.To()
				if !bytes.Equal(trigger.To[:], txTo[:]) {
					log.Trace().Int("index", i).Msg("To does not match; ignoring")
					continue
				}
			}
			trigger.Handler.HandleTx(ctx, tx, trigger)
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

		if to+1-from > maxBlocksForEvents {
			to = from + maxBlocksForEvents - 1
		}

		if err := s.pollEventsForTrigger(ctx, trigger, from, to); err != nil {
			return err
		}

		md.LatestBlocks[trigger.Name] = to
		if err := s.setEventsMetadata(ctx, md); err != nil {
			return errors.Wrap(err, "failed to set metadata after event poll")
		}
	}

	return nil
}

func (s *Service) pollEventsForTrigger(ctx context.Context,
	trigger *handlers.EventTrigger,
	from uint32,
	to uint32,
) error {
	// Resolve the source.
	var source *types.Address
	var err error
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

	return nil
}
