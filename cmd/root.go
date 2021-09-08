// Copyright (c) 2019 The Jaeger Authors.
// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/uber/jaeger-lib/metrics"
	jexpvar "github.com/uber/jaeger-lib/metrics/expvar"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/patil-kshitij/hotrod/services/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"stash.logicmonitor.com/~lalit.shukla_logicmonitor.com/opentelemetry-go-contrib.git/resource/detectors/aws/ec2"
)

var (
	metricsBackend string
	logger         *zap.Logger
	metricsFactory metrics.Factory

	fixDBConnDelay         time.Duration
	fixDBConnDisableMutex  bool
	fixRouteWorkerPoolSize int

	customerPort int
	driverPort   int
	frontendPort int
	routePort    int

	basepath string
	jaegerUI string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "examples-hotrod",
	Short: "HotR.O.D. - A tracing demo application",
	Long:  `HotR.O.D. - A tracing demo application.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logger.Fatal("We bowled a googly", zap.Error(err))
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&metricsBackend, "metrics", "m", "expvar", "Metrics backend (expvar|prometheus)")
	RootCmd.PersistentFlags().DurationVarP(&fixDBConnDelay, "fix-db-query-delay", "D", 300*time.Millisecond, "Average latency of MySQL DB query")
	RootCmd.PersistentFlags().BoolVarP(&fixDBConnDisableMutex, "fix-disable-db-conn-mutex", "M", false, "Disables the mutex guarding db connection")
	RootCmd.PersistentFlags().IntVarP(&fixRouteWorkerPoolSize, "fix-route-worker-pool-size", "W", 3, "Default worker pool size")

	// Add flags to choose ports for services
	RootCmd.PersistentFlags().IntVarP(&customerPort, "customer-service-port", "c", 8081, "Port for customer service")
	RootCmd.PersistentFlags().IntVarP(&driverPort, "driver-service-port", "d", 8082, "Port for driver service")
	RootCmd.PersistentFlags().IntVarP(&frontendPort, "frontend-service-port", "f", 8080, "Port for frontend service")
	RootCmd.PersistentFlags().IntVarP(&routePort, "route-service-port", "r", 8083, "Port for routing service")

	// Flag for serving frontend at custom basepath url
	RootCmd.PersistentFlags().StringVarP(&basepath, "basepath", "b", "", `Basepath for frontend service(default "/")`)
	RootCmd.PersistentFlags().StringVarP(&jaegerUI, "jaeger-ui", "j", "http://localhost:16686", "Address of Jaeger UI to create [find trace] links")

	rand.Seed(int64(time.Now().Nanosecond()))
	logger, _ = zap.NewDevelopment(
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(1),
	)
	cobra.OnInitialize(onInitialize)
}

// onInitialize is called before the command is executed.
func onInitialize() {

	initTracing()

	switch metricsBackend {
	case "expvar":
		metricsFactory = jexpvar.NewFactory(10) // 10 buckets for histograms
		logger.Info("Using expvar as metrics backend")
	case "prometheus":
		metricsFactory = jprom.New().Namespace(metrics.NSOptions{Name: "hotrod", Tags: nil})
		logger.Info("Using Prometheus as metrics backend")
	default:
		logger.Fatal("unsupported metrics backend " + metricsBackend)
	}
	if config.MySQLGetDelay != fixDBConnDelay {
		logger.Info("fix: overriding MySQL query delay", zap.Duration("old", config.MySQLGetDelay), zap.Duration("new", fixDBConnDelay))
		config.MySQLGetDelay = fixDBConnDelay
	}
	if fixDBConnDisableMutex {
		logger.Info("fix: disabling db connection mutex")
		config.MySQLMutexDisabled = true
	}
	if config.RouteWorkerPoolSize != fixRouteWorkerPoolSize {
		logger.Info("fix: overriding route worker pool size", zap.Int("old", config.RouteWorkerPoolSize), zap.Int("new", fixRouteWorkerPoolSize))
		config.RouteWorkerPoolSize = fixRouteWorkerPoolSize
	}

	if customerPort != 8081 {
		logger.Info("changing customer service port", zap.Int("old", 8081), zap.Int("new", customerPort))
	}

	if driverPort != 8082 {
		logger.Info("changing driver service port", zap.Int("old", 8082), zap.Int("new", driverPort))
	}

	if frontendPort != 8080 {
		logger.Info("changing frontend service port", zap.Int("old", 8080), zap.Int("new", frontendPort))
	}

	if routePort != 8083 {
		logger.Info("changing route service port", zap.Int("old", 8083), zap.Int("new", routePort))
	}

	if basepath != "" {
		logger.Info("changing basepath for frontend", zap.String("old", "/"), zap.String("new", basepath))
	}
}

func logError(logger *zap.Logger, err error) error {
	if err != nil {
		logger.Error("Error running command", zap.Error(err))
	}
	return err
}

func initTracing() {
	ctx := context.Background()

	//res := lmdetector.New(ctx)
	res, err := ec2.NewResourceDetector().Detect(ctx)
	if err != nil {
		fmt.Println("Error in detecting ec2 resource")
	}

	fmt.Println("Resource Detected :", res)

	traceExporter, _ := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint("localhost:55681"),
	)

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

}
