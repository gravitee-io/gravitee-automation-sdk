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
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-mock-server/server"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/auth"
	"github.com/spf13/cobra"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newCommand().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	var port int
	var basePath string
	var authFile string
	var dryRunReject bool

	cmd := &cobra.Command{
		Use:          "apim-mock-server",
		Short:        "Start the APIM automation mock server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var reg *auth.Registry
			if authFile != "" {
				cfg, err := auth.LoadConfig(authFile)
				if err != nil {
					return err
				}
				reg = auth.NewRegistry(*cfg, basePath)
			}
			return serve(cmd.Context(), port, basePath, reg, dryRunReject)
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "HTTP listen port")
	cmd.Flags().StringVar(&basePath, "base-path", server.BasePath, "API base path")
	cmd.Flags().StringVar(&authFile, "auth-file", "", "Path to auth config YAML (optional)")
	cmd.Flags().BoolVar(&dryRunReject, "dry-run-reject", false, "Reject PUT when ?dryRun=true")
	cmd.CompletionOptions.DisableDefaultCmd = true

	return cmd
}

func serve(ctx context.Context, port int, basePath string, reg *auth.Registry, dryRunReject bool) error {
	handler := server.NewWithPath(server.NewMockAPIM(dryRunReject), basePath, reg)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("APIM mock server listening on http://localhost:%d%s", port, basePath)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		return srv.Shutdown(context.Background())
	}
}
