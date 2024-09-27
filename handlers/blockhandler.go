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

package handlers

import (
	"context"

	"github.com/attestantio/go-execution-client/spec"
)

// BlockTrigger is a trigger for a block.
type BlockTrigger struct {
	Name          string
	EarliestBlock uint32
	Handler       BlockHandler
}

// BlockHandlerFunc defines the handler function.
type BlockHandlerFunc func(ctx context.Context, block *spec.Block, trigger *BlockTrigger)

// BlockHandler defines the methods that need to be implemented to handle block events.
type BlockHandler interface {
	// HandleBlock handles a block provided by the listener.
	// If this call returns an error then the listener will not send further blocks in the current poll,
	// and on the next poll it will start again with this block.
	HandleBlock(ctx context.Context, block *spec.Block, trigger *BlockTrigger) error
}
