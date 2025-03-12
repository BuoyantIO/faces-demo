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
	"log/slog"
	"os"
)

func InitLogging() {
	logLevel := &slog.LevelVar{} // INFO

	slogOpts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, slogOpts))
	slog.SetDefault(logger)

	if BoolFromEnv("DEBUG_ENABLED", false) {
		logLevel.Set(slog.LevelDebug)
	}
}

func Infof(format string, args ...interface{}) {
	slog.Info(fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	slog.Warn(fmt.Sprintf(format, args...))
}
