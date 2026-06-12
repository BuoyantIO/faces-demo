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
	"crypto/rand"
	"fmt"
)

// FaceMessage is the unit of data flowing through the pub/sub pipeline.
// Both the publisher and subscriber share this type; it is also the JSON
// payload stored in the Redis list.
type FaceMessage struct {
	ID        string   `json:"id"`
	Smiley    string   `json:"smiley"`
	Color     string   `json:"color"`
	Errors    []string `json:"errors,omitempty"`
	Timestamp int64    `json:"ts"`
}

// NewFaceMessageID generates a UUID v4 using crypto/rand with no external
// dependencies.  Bit positions follow RFC 4122 §4.4.
func NewFaceMessageID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits (RFC 4122)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
