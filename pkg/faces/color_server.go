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

package faces

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"fmt"

	"github.com/BuoyantIO/faces-demo/v2/pkg/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type colorServer struct {
	color.UnimplementedColorServiceServer
	provider *ColorProvider
}

func NewColorServer(provider *ColorProvider) *colorServer {
	return &colorServer{provider: provider}
}

func (srv *colorServer) Start(port int) error {
	var grpcOpts []grpc.ServerOption

	grpcServer := grpc.NewServer(grpcOpts...)
	color.RegisterColorServiceServer(grpcServer, srv)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	slog.Info(fmt.Sprintf("listening on %s", listener.Addr()))
	grpcServer.Serve(listener)

	return nil
}

func (srv *colorServer) BuildResponse(resp *ProviderResponse) (*color.ColorResponse, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		return &color.ColorResponse{
			Color: resp.GetString("color"),
		}, nil

	case http.StatusTooManyRequests:
		return nil, status.Errorf(codes.ResourceExhausted, "rate limited: %s", resp.GetErrors())

	default:
		return nil, status.Errorf(codes.Internal, "failed to get color: %s", resp.GetErrors())
	}
}

func (srv *colorServer) Center(ctx context.Context, req *color.ColorRequest) (*color.ColorResponse, error) {
	resp, err := HandleGRPC(ctx, &srv.provider.BaseProvider, "center", int(req.Row), int(req.Column))

	if err != nil {
		return nil, err
	}

	return srv.BuildResponse(resp)
}

func (srv *colorServer) Edge(ctx context.Context, req *color.ColorRequest) (*color.ColorResponse, error) {
	resp, err := HandleGRPC(ctx, &srv.provider.BaseProvider, "edge", int(req.Row), int(req.Column))

	if err != nil {
		return nil, err
	}

	return srv.BuildResponse(resp)
}
