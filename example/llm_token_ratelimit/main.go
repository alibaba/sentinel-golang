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

package main

import (
	"fmt"
	"llm_token_ratelimit/ratelimit"
	"llm_token_ratelimit/server"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
)

func StartMonitor() {
	fmt.Println("pprof is running on http://127.0.0.1:6060/debug/pprof/")
	fmt.Println(http.ListenAndServe("127.0.0.1:6060", nil))
}

func StartService() {
	gin.SetMode(gin.ReleaseMode)
	ratelimit.InitSentinel()
	server.StartServer("127.0.0.1", 9527)
}

func main() {
	go StartMonitor()
	StartService()
}
