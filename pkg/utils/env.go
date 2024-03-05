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
	"os"
	"strconv"
)

func BoolFromEnv(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)

	if err != nil {
		return defaultValue
	}

	return value
}

func IntFromEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)

	if err != nil {
		return defaultValue
	}

	return value
}

// PercentageFromEnv is just like IntFromEnv, but it makes certain that the
// value is between 0 and 100, inclusive.
func PercentageFromEnv(key string, defaultValue int) int {
	value := IntFromEnv(key, defaultValue)

	if value < 0 {
		value = 0
	}

	if value > 100 {
		value = 100
	}

	return value
}

func FloatFromEnv(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)

	if err != nil {
		return defaultValue
	}

	return value
}

func StringFromEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
