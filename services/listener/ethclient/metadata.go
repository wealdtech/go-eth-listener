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

var (
	blocksMetadataKey       = []byte("listener.ethclient.blocks")
	transactionsMetadataKey = []byte("listener.ethclient.transactions")
	eventsMetadataKey       = []byte("listener.ethclient.events")
)

type blocksMetadata struct {
	LatestBlock int32 `json:"latest_block"`
}

type transactionsMetadata struct {
	LatestBlock int32 `json:"latest_block"`
}

type eventsMetadata struct {
	LatestBlocks map[string]uint32 `json:"latest_blocks"`
}

func (s *Service) getBlocksMetadata(_ context.Context) (*blocksMetadata, error) {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return nil, errors.New("database closed")
	}

	data, closer, err := s.metadataDB.Get(blocksMetadataKey)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return &blocksMetadata{
				LatestBlock: -1,
			}, nil
		}

		return nil, errors.Wrap(err, "failed to get blocks metadata")
	}

	if err := closer.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close blocks metadata")
	}

	res := &blocksMetadata{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal blocks metadata")
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
		return errors.Wrap(err, "failed to marshal blocks metadata")
	}

	if err := s.metadataDB.Set(blocksMetadataKey, data, pebble.Sync); err != nil {
		return errors.Wrap(err, "failed to set blocks metadata")
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

		return nil, errors.Wrap(err, "failed to get transactions metadata")
	}

	if err := closer.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close transactions metadata")
	}

	res := &transactionsMetadata{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal transactions metadata")
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
		return errors.Wrap(err, "failed to marshal transactions metadata")
	}

	if err := s.metadataDB.Set(transactionsMetadataKey, data, pebble.Sync); err != nil {
		return errors.Wrap(err, "failed to set transactions metadata")
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
				LatestBlocks: map[string]uint32{},
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

func (s *Service) setEventsMetadata(_ context.Context, md *eventsMetadata) error {
	s.metadataDBMu.Lock()
	defer s.metadataDBMu.Unlock()
	if !s.metadataDBOpen.Load() {
		return errors.New("database closed")
	}

	data, err := json.Marshal(md)
	if err != nil {
		return errors.Wrap(err, "failed to marshal events metadata")
	}

	if err := s.metadataDB.Set(eventsMetadataKey, data, pebble.Sync); err != nil {
		return errors.Wrap(err, "failed to set events metadata")
	}

	return nil
}
