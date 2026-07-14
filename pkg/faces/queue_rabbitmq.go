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
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

// rabbitmqBackend implements QueueBackend using an AMQP queue.
//
// A single connection is shared; two long-lived channels are kept open — one
// for publishing (pushCh) and one for consuming (popCh).  Both operations are
// protected by mu so that the publisher's concurrent goroutines never race on
// the same channel (AMQP channels are not goroutine-safe).
//
// The queue is declared with x-max-length so that when the depth cap is
// reached RabbitMQ automatically drops the oldest messages (drop-head).  This
// mirrors the Redis LTRIM behaviour without requiring a Trim call on every Push.
type rabbitmqBackend struct {
	url       string
	queueName string
	maxDepth  int64
	logger    *slog.Logger

	mu     sync.Mutex
	conn   *amqp.Connection
	pushCh *amqp.Channel
	popCh  *amqp.Channel
}

func newRabbitMQBackendFromEnvironment(logger *slog.Logger) (*rabbitmqBackend, error) {
	host := utils.StringFromEnv("RABBITMQ_HOST", "rabbitmq")
	port := utils.StringFromEnv("RABBITMQ_PORT", "5672")
	user := utils.StringFromEnv("RABBITMQ_USER", "faces")
	password := utils.StringFromEnv("RABBITMQ_PASSWORD", "faces-password")
	vhost := utils.StringFromEnv("RABBITMQ_VHOST", "/")
	queueName := utils.StringFromEnv("RABBITMQ_QUEUE", "faces.queue")
	maxDepth := int64(utils.IntFromEnv("RABBITMQ_MAX_QUEUE_DEPTH", 5000))

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, port, vhost)

	b := &rabbitmqBackend{
		url:       url,
		queueName: queueName,
		maxDepth:  maxDepth,
		logger:    logger,
	}

	// Retry with exponential backoff — mirrors the MySQL connect pattern.
	// RabbitMQ may still be initialising even after the init container passes.
	delay := 2 * time.Second
	for {
		err := b.connect()
		if err == nil {
			break
		}
		logger.Warn("QueueBackend: RabbitMQ not ready", "error", err, "retry_in", delay.String())
		time.Sleep(delay)
		if delay < 30*time.Second {
			delay *= 2
		}
	}

	logger.Info("QueueBackend: RabbitMQ connected",
		"host", host, "queue", queueName, "maxDepth", maxDepth)
	return b, nil
}

// connect (re)establishes the AMQP connection and both channels, then
// declares the queue.  Called on startup and on reconnect.
func (b *rabbitmqBackend) connect() error {
	conn, err := amqp.Dial(b.url)
	if err != nil {
		return fmt.Errorf("rabbitmq: dial: %w", err)
	}

	pushCh, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("rabbitmq: open push channel: %w", err)
	}

	popCh, err := conn.Channel()
	if err != nil {
		pushCh.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq: open pop channel: %w", err)
	}

	// Declare with x-max-length so RabbitMQ enforces the depth cap automatically.
	// Using "drop-head" overflow so the oldest message is evicted when full —
	// equivalent to Redis LTRIM behaviour.
	args := amqp.Table{
		"x-max-length": b.maxDepth,
		"x-overflow":   "drop-head",
	}
	if _, err := pushCh.QueueDeclare(b.queueName, true, false, false, false, args); err != nil {
		pushCh.Close()
		popCh.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq: declare queue %q: %w", b.queueName, err)
	}

	b.conn = conn
	b.pushCh = pushCh
	b.popCh = popCh
	return nil
}

// ensureConnected reconnects if the connection is closed.  Must be called
// while holding mu.
func (b *rabbitmqBackend) ensureConnected() error {
	if b.conn != nil && !b.conn.IsClosed() {
		return nil
	}
	b.logger.Warn("QueueBackend: RabbitMQ reconnecting")
	return b.connect()
}

func (b *rabbitmqBackend) Push(ctx context.Context, msg *FaceMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("rabbitmq push: marshal: %w", err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.ensureConnected(); err != nil {
		return err
	}

	err = b.pushCh.PublishWithContext(ctx, "", b.queueName, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		// Mark connection dead so the next call triggers a reconnect.
		b.conn = nil
	}
	return err
}

func (b *rabbitmqBackend) Pop(ctx context.Context) (*FaceMessage, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.ensureConnected(); err != nil {
		return nil, err
	}

	delivery, ok, err := b.popCh.Get(b.queueName, true /* autoAck */)
	if err != nil {
		b.conn = nil
		return nil, fmt.Errorf("rabbitmq pop: %w", err)
	}
	if !ok {
		return nil, nil // queue empty
	}

	var msg FaceMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return nil, fmt.Errorf("rabbitmq pop: unmarshal: %w", err)
	}
	return &msg, nil
}

// Warm publishes msgs oldest-first.  AMQP queues are FIFO so the first
// published message is the first consumed — no reverse-order trick needed.
func (b *rabbitmqBackend) Warm(ctx context.Context, msgs []*FaceMessage) error {
	for _, msg := range msgs {
		if err := b.Push(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Depth returns the current number of messages in the queue using a passive
// QueueInspect (no side effects).  Called only by the publisher's monitor
// goroutine; protected by the existing mutex.
func (b *rabbitmqBackend) Depth(ctx context.Context) (int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.ensureConnected(); err != nil {
		return 0, err
	}

	q, err := b.pushCh.QueueInspect(b.queueName)
	if err != nil {
		b.conn = nil // force reconnect on next call
		return 0, fmt.Errorf("rabbitmq depth: %w", err)
	}
	return int64(q.Messages), nil
}

func (b *rabbitmqBackend) MaxDepth() int64 {
	return b.maxDepth
}

func (b *rabbitmqBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.pushCh != nil {
		b.pushCh.Close()
	}
	if b.popCh != nil {
		b.popCh.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
