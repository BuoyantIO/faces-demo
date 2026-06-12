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
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

// FaceStore is the MySQL persistence layer for the pub/sub pipeline.
// All SQL is confined to this file; neither publisher nor subscriber
// import database/sql directly.
type FaceStore struct {
	db              *sql.DB
	logger          *slog.Logger
	retentionPeriod time.Duration
}

// NewFaceStoreFromEnvironment reads DB_* env vars, opens a connection pool,
// pings MySQL, and runs the schema migration.  Both publisher and subscriber
// call this on startup; if it returns an error the caller should panic.
func NewFaceStoreFromEnvironment() (*FaceStore, error) {
	logger := slog.Default().With("component", "FaceStore")

	dsn := utils.StringFromEnv("DB_DSN", "")
	if dsn == "" {
		host := utils.StringFromEnv("DB_HOST", "mysql")
		port := utils.StringFromEnv("DB_PORT", "3306")
		dbName := utils.StringFromEnv("DB_NAME", "faces")
		user := utils.StringFromEnv("DB_USER", "faces")
		password := utils.StringFromEnv("DB_PASSWORD", "")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
			user, password, host, port, dbName)
	}

	maxOpen := utils.IntFromEnv("DB_MAX_OPEN", 10)
	maxIdle := utils.IntFromEnv("DB_MAX_IDLE", 5)
	retentionHours := utils.IntFromEnv("DB_RETENTION_H", 24)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("FaceStore: sql.Open: %w", err)
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Retry until MySQL is reachable, logging each failure. This keeps the
	// pod in Running state rather than crashing while the database initialises.
	delay := 2 * time.Second
	for {
		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr := db.PingContext(pingCtx)
		cancel()
		if pingErr == nil {
			break
		}
		logger.Warn("FaceStore: waiting for MySQL", "error", pingErr, "retry_in", delay.String())
		time.Sleep(delay)
		if delay < 30*time.Second {
			delay *= 2
		}
	}

	fs := &FaceStore{
		db:              db,
		logger:          logger,
		retentionPeriod: time.Duration(retentionHours) * time.Hour,
	}

	if err := fs.migrate(); err != nil {
		return nil, fmt.Errorf("FaceStore: migrate: %w", err)
	}

	logger.Info("FaceStore: connected and migrated")
	return fs, nil
}

// migrate creates the schema and applies any in-place column migrations.
// Safe to call on every restart — all operations are idempotent.
//
// State lifecycle:
//   pending      → INSERT happened; queue push not yet attempted (crash-recovery window, ~ms)
//   queued       → Successfully pushed to queue; waiting for subscriber to consume
//   acknowledged → Consumed by face-subscriber (GUI actually displayed this face)
func (fs *FaceStore) migrate() error {
	// Create table on fresh installs.
	if _, err := fs.db.Exec(`CREATE TABLE IF NOT EXISTS face_queue (
		id               VARCHAR(36)  NOT NULL,
		smiley           TEXT         NOT NULL,
		color            TEXT         NOT NULL,
		errors           TEXT         NULL,
		state            ENUM('pending','queued','acknowledged') NOT NULL DEFAULT 'pending',
		created_at       BIGINT       NOT NULL,
		acknowledged_at  BIGINT       NULL,
		PRIMARY KEY (id),
		INDEX idx_state_created (state, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}

	// Migrate existing installs that only have the old two-value ENUM.
	// This ALTER is idempotent: MySQL accepts it silently if 'queued' already exists.
	_, _ = fs.db.Exec(`ALTER TABLE face_queue
		MODIFY COLUMN state ENUM('pending','queued','acknowledged') NOT NULL DEFAULT 'pending'`)

	return nil
}

// Insert writes a new message with state='pending'.  Called by the publisher
// before pushing to Redis (write-ahead guarantee).
func (fs *FaceStore) Insert(ctx context.Context, msg *FaceMessage) error {
	var errJSON interface{}
	if len(msg.Errors) > 0 {
		b, _ := json.Marshal(msg.Errors)
		errJSON = string(b)
	}

	_, err := fs.db.ExecContext(ctx,
		`INSERT INTO face_queue (id, smiley, color, errors, state, created_at)
		 VALUES (?, ?, ?, ?, 'pending', ?)`,
		msg.ID, msg.Smiley, msg.Color, errJSON, msg.Timestamp,
	)
	return err
}

// UnacknowledgedMessages returns all rows that have not yet been delivered to
// the GUI, ordered oldest-first.  This includes both:
//   - 'pending'  — INSERT happened but queue push not yet attempted (crash window)
//   - 'queued'   — was successfully pushed, but may no longer be in the queue
//                  (e.g. after a manual purge or queue pod restart)
//
// Called by WarmQueue on publisher startup and on demand via /control.
// Re-pushing 'queued' rows is safe: if the message is already in the queue,
// the subscriber processes it a second time; the MySQL ack is idempotent.
func (fs *FaceStore) UnacknowledgedMessages(ctx context.Context) ([]*FaceMessage, error) {
	rows, err := fs.db.QueryContext(ctx,
		`SELECT id, smiley, color, errors, created_at
		 FROM   face_queue
		 WHERE  state IN ('pending', 'queued')
		 ORDER  BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*FaceMessage
	for rows.Next() {
		var msg FaceMessage
		var errJSON sql.NullString
		if err := rows.Scan(&msg.ID, &msg.Smiley, &msg.Color, &errJSON, &msg.Timestamp); err != nil {
			return nil, err
		}
		if errJSON.Valid && errJSON.String != "" && errJSON.String != "null" {
			_ = json.Unmarshal([]byte(errJSON.String), &msg.Errors)
		}
		msgs = append(msgs, &msg)
	}
	return msgs, rows.Err()
}

// MarkQueued transitions a row from 'pending' to 'queued' after it has been
// successfully pushed to the message queue.  Called by the publisher immediately
// after a successful Push.  Idempotent — no-op if already 'queued'.
func (fs *FaceStore) MarkQueued(ctx context.Context, id string) error {
	_, err := fs.db.ExecContext(ctx,
		`UPDATE face_queue SET state = 'queued' WHERE id = ? AND state = 'pending'`,
		id,
	)
	return err
}

// Acknowledge marks a row as consumed by the subscriber.  The guard on
// state IN ('pending','queued') makes this idempotent.
func (fs *FaceStore) Acknowledge(ctx context.Context, id string) error {
	_, err := fs.db.ExecContext(ctx,
		`UPDATE face_queue
		 SET    state = 'acknowledged',
		        acknowledged_at = ?
		 WHERE  id = ?
		   AND  state IN ('pending','queued')`,
		time.Now().UnixMilli(), id,
	)
	return err
}

// DeleteOldRows removes rows older than the configured retention period,
// regardless of state. Acknowledged rows are cleaned up normally. Pending rows
// older than the retention period are stranded — they were dropped by the
// queue depth cap (Redis LTRIM / RabbitMQ x-max-length) and will never be
// acknowledged, so they are safe to delete.
func (fs *FaceStore) DeleteOldRows(ctx context.Context) (int64, error) {
	cutoff := time.Now().Add(-fs.retentionPeriod).UnixMilli()
	res, err := fs.db.ExecContext(ctx,
		`DELETE FROM face_queue WHERE created_at < ?`,
		cutoff,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// QueueDepth returns per-state row counts for health checks and the admin
// /control endpoint.
//
//   pending      ≈ 0 under normal conditions (only during crash-recovery window)
//   queued       = messages pushed to queue but not yet consumed by subscriber
//   acknowledged = messages the subscriber has actually delivered to the GUI
func (fs *FaceStore) QueueDepth(ctx context.Context) (pending, queued, acknowledged int64, err error) {
	row := fs.db.QueryRowContext(ctx,
		`SELECT
			SUM(CASE WHEN state='pending'      THEN 1 ELSE 0 END),
			SUM(CASE WHEN state='queued'       THEN 1 ELSE 0 END),
			SUM(CASE WHEN state='acknowledged' THEN 1 ELSE 0 END)
		 FROM face_queue`,
	)
	err = row.Scan(&pending, &queued, &acknowledged)
	return
}
