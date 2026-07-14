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
	context "context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func grpcMetadata(ctx context.Context, prv *BaseProvider) (string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return "", "", status.Errorf(codes.DataLoss, "failed to get metadata")
	}

	user := ""
	users := md.Get(prv.GetUserHeaderName())

	if len(users) > 0 {
		user = users[0]
	}

	userAgent := ""
	userAgents := md.Get("user-agent")

	if len(userAgents) > 0 {
		userAgent = userAgents[0]
	}

	return user, userAgent, nil
}

func HandleGRPC(ctx context.Context, prv *BaseProvider, subrequest string, row, col int) (*ProviderResponse, error) {
	start := time.Now()
	user, userAgent, err := grpcMetadata(ctx, prv)

	if err != nil {
		return nil, status.Errorf(codes.DataLoss, "failed to get user")
	}

	prvReq := &ProviderRequest{
		subrequest: subrequest,
		user:       user,
		userAgent:  userAgent,
		row:        row,
		col:        col,
	}

	resp := prv.HandleRequest(start, prvReq)

	return &resp, nil
}
