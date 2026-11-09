// Copyright (C) 2015 The Gravitee team (http://gravitee.io)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gravitee-io/gravitee-automation-sdk/am-mock-server/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_DefaultFlags(t *testing.T) {
	cmd := newCommand()

	port, err := cmd.Flags().GetInt("port")
	require.NoError(t, err)
	assert.Equal(t, 8080, port)

	basePath, err := cmd.Flags().GetString("base-path")
	require.NoError(t, err)
	assert.Equal(t, server.BasePath, basePath)

	dryRunReject, err := cmd.Flags().GetBool("dry-run-reject")
	require.NoError(t, err)
	assert.False(t, dryRunReject)
}

func TestCommand_ParseFlags(t *testing.T) {
	cmd := newCommand()

	err := cmd.ParseFlags([]string{"--port", "9090", "--base-path", "/api", "--dry-run-reject"})
	require.NoError(t, err)

	port, err := cmd.Flags().GetInt("port")
	require.NoError(t, err)
	assert.Equal(t, 9090, port)

	basePath, err := cmd.Flags().GetString("base-path")
	require.NoError(t, err)
	assert.Equal(t, "/api", basePath)

	dryRunReject, err := cmd.Flags().GetBool("dry-run-reject")
	require.NoError(t, err)
	assert.True(t, dryRunReject)
}

func TestServe_AnswersTheFaultsEndpoint(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, port, server.BasePath, nil, false) }()

	assert.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", port, server.FaultsPath))
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 20*time.Millisecond)

	cancel()
	require.NoError(t, <-done)
}
