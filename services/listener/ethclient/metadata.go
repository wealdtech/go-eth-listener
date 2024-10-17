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
	"errors"

	"github.com/cockroachdb/pebble"
)

var (
	blocksMetadataKey       = []byte("listener.ethclient.blocks")
	transactionsMetadataKey = []byte("listener.ethclient.transactions")
	eventsMetadataKey       = []byte("listener.ethclient.events")
)

type blocksMetadata struct {
	LatestBlocks map[string]int32 `json:"latest_blocks"`
}

type transactionsMetadata struct {
	LatestBlock int32 `json:"latest_block"`
}

type eventsMetadata struct {
	// LatestBlocks is deprecated.
	LatestBlocks map[string]uint32               `json:"latest_blocks,omitempty"`
	Entries      map[string]*eventsEntryMetadata `json:"entries"`
}

type eventsEntryMetadata struct {
	LatestBlock      uint32 `json:"latest_block"`
	LatestEventIndex int32  `json:"latest_event_index"`
}

func (s *Service) getBlocksMetadata(_ context.Context) (*blocksMetadata, error) {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return nil, errors.New("database closed")
	}

	res := &blocksMetadata{
		LatestBlocks: map[string]int32{},
	}

	data, closer, err := s.metadataDB.Get(blocksMetadataKey)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return res, nil
		}

		return nil, errors.Join(errors.New("failed to get blocks metadata"), err)
	}

	if err := closer.Close(); err != nil {
		return nil, errors.Join(errors.New("failed to close blocks metadata"), err)
	}

	if err := json.Unmarshal(data, res); err != nil {
		return nil, errors.Join(errors.New("failed to unmarshal blocks metadata"), err)
	}

	return res, nil
}

func (s *Service) setBlocksMetadata(_ context.Context, md *blocksMetadata) error {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return errors.New("database closed")
	}

	data, err := json.Marshal(md)
	if err != nil {
		return errors.Join(errors.New("failed to marshal blocks metadata"), err)
	}

	if err := s.metadataDB.Set(blocksMetadataKey, data, pebble.Sync); err != nil {
		return errors.Join(errors.New("failed to set blocks metadata"), err)
	}

	return nil
}

func (s *Service) getTransactionsMetadata(_ context.Context) (*transactionsMetadata, error) {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return nil, errors.New("database closed")
	}

	data, closer, err := s.metadataDB.Get(transactionsMetadataKey)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return &transactionsMetadata{
				LatestBlock: -1,
			}, nil
		}

		return nil, errors.Join(errors.New("failed to get transactions metadata"), err)
	}

	if err := closer.Close(); err != nil {
		return nil, errors.Join(errors.New("failed to close transactions metadata"), err)
	}

	res := &transactionsMetadata{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, errors.Join(errors.New("failed to unmarshal transactions metadata"), err)
	}

	return res, nil
}

func (s *Service) setTransactionsMetadata(_ context.Context, md *transactionsMetadata) error {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return errors.New("database closed")
	}

	data, err := json.Marshal(md)
	if err != nil {
		return errors.Join(errors.New("failed to marshal transactions metadata"), err)
	}

	if err := s.metadataDB.Set(transactionsMetadataKey, data, pebble.Sync); err != nil {
		return errors.Join(errors.New("failed to set transactions metadata"), err)
	}

	return nil
}

func (s *Service) getEventsMetadata(_ context.Context) (*eventsMetadata, error) {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return nil, errors.New("database closed")
	}

	data, closer, err := s.metadataDB.Get(eventsMetadataKey)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return &eventsMetadata{
				Entries: map[string]*eventsEntryMetadata{},
			}, nil
		}

		return nil, errors.Join(errors.New("failed to get events metadata"), err)
	}

	if err := closer.Close(); err != nil {
		return nil, errors.Join(errors.New("failed to close events metadata"), err)
	}

	res := &eventsMetadata{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, errors.Join(errors.New("failed to unmarshal events metadata"), err)
	}

	if res.Entries == nil {
		res.Entries = map[string]*eventsEntryMetadata{}
		for k, v := range res.LatestBlocks {
			res.Entries[k] = &eventsEntryMetadata{
				LatestBlock:      v,
				LatestEventIndex: -1,
			}
		}
		res.LatestBlocks = nil
	}

	return res, nil
}

func (s *Service) setEventsMetadata(_ context.Context, md *eventsMetadata) error {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return errors.New("database closed")
	}

	data, err := json.Marshal(md)
	if err != nil {
		return errors.Join(errors.New("failed to marshal events metadata"), err)
	}

	if err := s.metadataDB.Set(eventsMetadataKey, data, pebble.Sync); err != nil {
		return errors.Join(errors.New("failed to set events metadata"), err)
	}

	return nil
}
