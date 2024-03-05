// SPDX-FileCopyrightText: 2024 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2024 Buoyant Inc.
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

package utils

import (
	"fmt"
	"sync"
	"time"
)

// RateCounter counts events in an N-second window and averages them. It does this
// by maintaining a (thread-safe) set of buckets, one per second, and providing a way
// to increment the bucket's counters.

type RateCounter struct {
	numberOfBuckets int
	firstBucket     time.Time
	buckets         []int
	lock            sync.Mutex
}

func NewRateCounter(numberOfBuckets int) *RateCounter {
	rc := &RateCounter{
		numberOfBuckets: numberOfBuckets,
		buckets:         make([]int, numberOfBuckets),
	}

	// Start a timer to tick the RateCounter every second, so that
	// rates fall off over time if nothing is happening.
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			rc.Tick(time.Now())
		}
	}()

	return rc
}

func (rc *RateCounter) String() string {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	return fmt.Sprintf("RateCounter@%v: %v", rc.firstBucket, rc.buckets)
}

// CurrentRate returns the current rate as a float.
func (rc *RateCounter) CurrentRate() float64 {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	sum := 0
	for _, count := range rc.buckets {
		sum += count
	}
	return float64(sum) / float64(rc.numberOfBuckets)
}

// Tick marks the passage of time. It can shift windows, but it can't actually
// record an event. It returns the current bucket number.
func (rc *RateCounter) Tick(now time.Time) int {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if rc.firstBucket.IsZero() {
		rc.firstBucket = now
	}

	bucket := int(now.Sub(rc.firstBucket).Seconds())

	if bucket >= rc.numberOfBuckets {
		// We've moved past the end of the buckets, so slide the whole
		// window over.
		numberPast := bucket - rc.numberOfBuckets + 1

		rc.firstBucket = rc.firstBucket.Add(time.Duration(numberPast) * time.Second)

		if numberPast >= rc.numberOfBuckets {
			rc.buckets = make([]int, rc.numberOfBuckets)
		} else {
			rc.buckets = append(rc.buckets[numberPast:], make([]int, numberPast)...)
		}

		bucket = int(now.Sub(rc.firstBucket).Seconds())
	}

	return bucket
}

// Mark records that a request has happened. It's a Tick plus incrementing the
// current bucket.
func (rc *RateCounter) Mark(now time.Time) {
	bucket := rc.Tick(now)
	rc.buckets[bucket]++
}
