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

package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	"flag"
	"fmt"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type colorServer struct {
	faces.UnimplementedColorServiceServer
	provider *faces.ColorProvider
}

func (srv *colorServer) Center(ctx context.Context, req *faces.ColorRequest) (*faces.ColorResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}

	user := md.Get("user")

	baseResp := srv.provider.Get(int(req.Row), int(req.Column))

	slog.Debug(fmt.Sprintf("CENTER: %d, %d (%s) => %d, %s\n", req.Row, req.Column, user, baseResp.StatusCode, baseResp.Body))

	switch baseResp.StatusCode {
	case http.StatusOK:
		color := baseResp.Body

		return &faces.ColorResponse{
			Color: color,
		}, nil

	case http.StatusTooManyRequests:
		return nil, status.Errorf(codes.ResourceExhausted, "rate limited: %s", baseResp.Body)

	default:
		return nil, status.Errorf(codes.Internal, "failed to get color: %s", baseResp.Body)
	}
}

func (srv *colorServer) Edge(ctx context.Context, req *faces.ColorRequest) (*faces.ColorResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}

	user := ""
	users := md.Get("x-faces-user")

	if len(users) > 0 {
		user = users[0]
	}

	baseResp := srv.provider.Get(int(req.Row), int(req.Column))

	slog.Debug(fmt.Sprintf("EDGE: %d, %d (%s) => %d, %s\n", req.Row, req.Column, user, baseResp.StatusCode, baseResp.Body))

	switch baseResp.StatusCode {
	case http.StatusOK:
		color := baseResp.Body

		return &faces.ColorResponse{
			Color: color,
		}, nil

	case http.StatusTooManyRequests:
		return nil, status.Errorf(codes.ResourceExhausted, "rate limited: %s", baseResp.Body)

	default:
		return nil, status.Errorf(codes.Internal, "failed to get color: %s", baseResp.Body)
	}
}

func main() {
	logLevel := &slog.LevelVar{} // INFO

	slogOpts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, slogOpts))
	slog.SetDefault(logger)

	logLevel.Set(slog.LevelDebug)

	// Define a command-line flag for the port number
	port := flag.Int("port", 8000, "the port number to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		slog.Error(fmt.Sprintf("failed to listen: %v", err))
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("listening on %s", listener.Addr()))
	var grpcOpts []grpc.ServerOption

	cprv := faces.NewColorProviderFromEnvironment()
	server := &colorServer{provider: cprv}

	grpcServer := grpc.NewServer(grpcOpts...)
	faces.RegisterColorServiceServer(grpcServer, server)
	grpcServer.Serve(listener)
}
