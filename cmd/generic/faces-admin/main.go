// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	utils.InitLogging()

	port := flag.Int("port", 8888, "the port number to listen on")
	flag.Parse()

	slog.Info("faces-admin starting", "version", version, "commit", commit, "date", date)

	admin := faces.NewAdminProviderFromEnvironment()
	admin.StartBackgroundPoller()
	admin.StartEmojivotoPoller()

	if err := admin.Start(fmt.Sprintf(":%d", *port)); err != nil {
		slog.Error("unable to serve HTTP", "error", err)
		os.Exit(1)
	}
}
