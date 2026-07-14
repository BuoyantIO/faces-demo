// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2025 Buoyant Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.  You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package faces

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

// QueueBackend abstracts the message queue used between the publisher and
// subscriber.  Both Redis (list-based) and RabbitMQ (AMQP) implementations
// satisfy this interface.  All methods must be safe for concurrent Push calls
// — the publisher spawns multiple goroutines.
type QueueBackend interface {
	// Push serialises msg and appends it to the tail of the queue.
	Push(ctx context.Context, msg *FaceMessage) error

	// Pop removes and returns the oldest message.
	// Returns (nil, nil) when the queue is empty — not an error.
	Pop(ctx context.Context) (*FaceMessage, error)

	// Warm bulk-loads msgs into the queue in oldest-first order.
	// Called on publisher startup and on demand to re-hydrate from MySQL.
	Warm(ctx context.Context, msgs []*FaceMessage) error

	// Depth returns the current number of messages in the queue.
	// Used by the publisher's queue monitor to detect under-fill.
	Depth(ctx context.Context) (int64, error)

	// MaxDepth returns the configured depth cap for this backend.
	// Used by the queue monitor to compute the fill threshold.
	MaxDepth() int64

	// Close releases any held connections or channels.
	Close() error
}

// NewQueueBackendFromEnvironment reads QUEUE_BACKEND ("rabbitmq" or "redis")
// and constructs the appropriate implementation.  Panics on unknown backend.
func NewQueueBackendFromEnvironment(logger *slog.Logger) (QueueBackend, error) {
	backend := utils.StringFromEnv("QUEUE_BACKEND", "rabbitmq")
	switch backend {
	case "redis":
		return newRedisBackendFromEnvironment(logger)
	case "rabbitmq":
		return newRabbitMQBackendFromEnvironment(logger)
	default:
		return nil, fmt.Errorf("unknown QUEUE_BACKEND %q: must be \"redis\" or \"rabbitmq\"", backend)
	}
}
