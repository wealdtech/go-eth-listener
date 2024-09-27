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
	"github.com/attestantio/go-execution-client/types"
)

// EventTrigger is a trigger for an event.
type EventTrigger struct {
	Name string
	// Source is a static address to use for event addresses.
	Source *types.Address
	// SourceResolver is a dynamic resolver use for event addresses.
	SourceResolver SourceResolver
	Topics         []types.Hash
	EarliestBlock  uint32
	Handler        EventHandler
}

// SourceResolver defines the methods that need to be implemented to resolve sources.
type SourceResolver interface {
	// Resolve resolves a source for events.
	Resolve(ctx context.Context) (*types.Address, error)
}

// EventHandlerFunc defines the handler function.
type EventHandlerFunc func(ctx context.Context, event *spec.BerlinTransactionEvent, trigger *EventTrigger)

// EventHandler defines the methods that need to be implemented to handle events.
type EventHandler interface {
	HandleEvent(ctx context.Context, event *spec.BerlinTransactionEvent, trigger *EventTrigger)
}
