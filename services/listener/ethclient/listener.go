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

	"github.com/attestantio/go-execution-client/api"
	executil "github.com/attestantio/go-execution-client/util"
	"github.com/rs/zerolog/log"
)

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
	md, err := s.getEventsMetadata(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get metadata for event poll")
		return
	}

	chainHeight, err := s.chainHeightProvider.ChainHeight(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get chain height for event poll")
		return
	}

	// Subtract the block delay.
	to := chainHeight - s.blockDelay

	for _, eventTrigger := range s.eventTriggers {
		// Obtain the last block we examined for this trigger, or
		// use the earliest block as defined in the trigger.
		from := eventTrigger.EarliestBlock
		if _, exists := md.LatestBlocks[eventTrigger.Name]; exists {
			next := md.LatestBlocks[eventTrigger.Name] + 1
			if next > from {
				from = next
			}
		}
		if from > to {
			s.log.Trace().Str("trigger", eventTrigger.Name).Uint32("from", from).Uint32("to", to).Msg("Not fetching events")
			return
		}
		// TODO maximum number of blocks to fetch.
		if to+1-from > 100 {
			to = from + 99
		}
		s.log.Trace().Str("trigger", eventTrigger.Name).Uint32("from", from).Uint32("to", to).Msg("Fetching events")

		filter := &api.EventsFilter{
			FromBlock: executil.MarshalUint32(from),
			ToBlock:   executil.MarshalUint32(to),
			Address:   eventTrigger.Source,
			Topics:    eventTrigger.Topics,
		}
		events, err := s.eventsProvider.Events(ctx, filter)
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to obtain events")
		}
		for _, event := range events {
			eventTrigger.Handler.HandleEvent(event, eventTrigger)
		}

		md.LatestBlocks[eventTrigger.Name] = to
		if err := s.setEventsMetadata(ctx, md); err != nil {
			log.Error().Err(err).Msg("Failed to set metadata after event poll")
		}
	}
}
