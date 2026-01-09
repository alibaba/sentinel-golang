// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package trpc provides Sentinel integration for tRPC-Go framework.
//
// For server side:
//
//	import (
//	    sentinelTrpc "github.com/alibaba/sentinel-golang/pkg/adapters/trpc"
//	    "trpc.group/trpc-go/trpc-go/filter"
//	)
//
//	// Register Sentinel filter globally
//	filter.Register("sentinel", sentinelTrpc.SentinelServerFilter(), nil)
//
//	// Or use it per-service in trpc_go.yaml:
//	// server:
//	//   filter:
//	//     - sentinel
//
// For client side:
//
//	import (
//	    sentinelTrpc "github.com/alibaba/sentinel-golang/pkg/adapters/trpc"
//	    "trpc.group/trpc-go/trpc-go/filter"
//	)
//
//	// Register Sentinel filter globally
//	filter.Register("sentinel", nil, sentinelTrpc.SentinelClientFilter())
//
//	// Or use it per-client in trpc_go.yaml:
//	// client:
//	//   filter:
//	//     - sentinel
//
// For both server and client:
//
//	import (
//	    sentinelTrpc "github.com/alibaba/sentinel-golang/pkg/adapters/trpc"
//	    "trpc.group/trpc-go/trpc-go/filter"
//	)
//
//	// Register Sentinel filter for both server and client
//	filter.Register("sentinel", sentinelTrpc.SentinelServerFilter(), sentinelTrpc.SentinelClientFilter())
//
// Custom resource extraction and block fallback:
//
//	// Custom resource extractor
//	serverFilter := sentinelTrpc.SentinelServerFilter(
//	    sentinelTrpc.WithResourceExtract(func(ctx context.Context, req interface{}) string {
//	        msg := codec.Message(ctx)
//	        return msg.ServerRPCName()
//	    }),
//	    sentinelTrpc.WithBlockFallback(func(ctx context.Context, req interface{}, blockErr *base.BlockError) (interface{}, error) {
//	        // Custom block handling logic
//	        return nil, errs.New(errs.RetServerSystemErr, "request blocked by sentinel")
//	    }),
//	)
package trpc
