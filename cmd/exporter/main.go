// Copyright 2025 Tobias Hintze
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
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/paraopsde/go-x/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/thz/dimplex-modbus-exporter/pkg/collector"
	"github.com/thz/dimplex-modbus-exporter/pkg/modbus"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "dimplex-modbus-exporter",
		Run: func(cmd *cobra.Command, args []string) {
			listen, _ := cmd.Flags().GetString("listen")
			addr, _ := cmd.Flags().GetString("address")

			ctx := context.Background()
			log := util.NewLogger()
			ctx = util.CtxWithLog(ctx, log)

			if err := run(ctx, addr, listen); err != nil {
				log.Error("failed to execute command", zap.Error(err))
				os.Exit(1)
			}
		},
	}
	rootCmd.Flags().StringP("listen", "l", ":9000", "listen address")
	rootCmd.Flags().StringP("address", "a", "", "address of the modbus server")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, addr, listen string) error {
	m, err := modbus.New(addr)
	if err != nil {
		return fmt.Errorf("failed to created modbus client: %w", err)
	}
	err = m.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to modbus: %w", err)
	}

	_, err = m.RequestPseudoFloat16(ctx, 6)
	if err != nil {
		return fmt.Errorf("failed to request data: %w", err)
	}

	collector := collector.New(m, ctx)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}
	return nil
}
