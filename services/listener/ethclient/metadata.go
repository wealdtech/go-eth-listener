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
	"encoding/json"

	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
)

var eventsMetadataKey = []byte("listener.ethclient.events")

type eventsMetadata struct {
	LatestBlocks map[string]uint32 `json:"latest_blocks"`
}

func (s *Service) getEventsMetadata(ctx context.Context) (*eventsMetadata, error) {
	data, closer, err := s.metadataDB.Get(eventsMetadataKey)
	if err != nil {
		if err == pebble.ErrNotFound {
			// This happens on first start, it's fine.
			return &eventsMetadata{
				LatestBlocks: map[string]uint32{
					"Staking deposits": 8674313,
				},
			}, nil
		}
		return nil, errors.Wrap(err, "failed to get events metadata")
	}

	if err := closer.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close events metadata")
	}

	res := &eventsMetadata{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal events metadata")
	}

	return res, nil
}

func (s *Service) setEventsMetadata(ctx context.Context, md *eventsMetadata) error {
	data, err := json.Marshal(md)
	if err != nil {
		return errors.Wrap(err, "failed to marshal events metadata")
	}

	if err := s.metadataDB.Set(eventsMetadataKey, data, pebble.Sync); err != nil {
		return errors.Wrap(err, "failed to set events metadata")
	}

	return nil
}
