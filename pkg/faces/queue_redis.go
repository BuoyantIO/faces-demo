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
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
	"github.com/redis/go-redis/v9"
)

// redisBackend implements QueueBackend using a Redis list.
// LPUSH adds to the head; RPOP removes from the tail — giving FIFO order.
// LTRIM on every Push caps the list at maxDepth, dropping the oldest entries.
type redisBackend struct {
	client   *redis.Client
	key      string
	maxDepth int64
	logger   *slog.Logger
}

func newRedisBackendFromEnvironment(logger *slog.Logger) (*redisBackend, error) {
	addr := utils.StringFromEnv("REDIS_ADDRESS", "redis:6379")
	password := utils.StringFromEnv("REDIS_PASSWORD", "")
	db := utils.IntFromEnv("REDIS_DB", 0)
	key := utils.StringFromEnv("REDIS_QUEUE_KEY", "faces:queue")
	maxDepth := int64(utils.IntFromEnv("REDIS_MAX_QUEUE_DEPTH", 5000))

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	logger.Info("QueueBackend: Redis", "addr", addr, "key", key, "maxDepth", maxDepth)
	return &redisBackend{client: client, key: key, maxDepth: maxDepth, logger: logger}, nil
}

func (r *redisBackend) Push(ctx context.Context, msg *FaceMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("redis push: marshal: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.LPush(ctx, r.key, string(payload))
	pipe.LTrim(ctx, r.key, 0, r.maxDepth-1)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisBackend) Pop(ctx context.Context) (*FaceMessage, error) {
	payload, err := r.client.RPop(ctx, r.key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var msg FaceMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		return nil, fmt.Errorf("redis pop: unmarshal: %w", err)
	}
	return &msg, nil
}

// Warm pushes msgs in reverse so that after LPUSH the oldest message ends up
// at the tail (right) where RPOP consumes it first — preserving FIFO order.
func (r *redisBackend) Warm(ctx context.Context, msgs []*FaceMessage) error {
	for i := len(msgs) - 1; i >= 0; i-- {
		if err := r.Push(ctx, msgs[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *redisBackend) Depth(ctx context.Context) (int64, error) {
	return r.client.LLen(ctx, r.key).Result()
}

func (r *redisBackend) MaxDepth() int64 {
	return r.maxDepth
}

func (r *redisBackend) Close() error {
	return r.client.Close()
}
