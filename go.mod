module github.com/patil-kshitij/hotrod

go 1.16

require (
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/jaegertracing/jaeger v1.26.0
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.2.1
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.22.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0 // indirect
	go.opentelemetry.io/otel v1.0.0-RC3
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.0.0-RC3
	go.opentelemetry.io/otel/sdk v1.0.0-RC3
	go.opentelemetry.io/otel/trace v1.0.0-RC3
	go.uber.org/zap v1.19.0
	google.golang.org/grpc v1.40.0
	stash.logicmonitor.com/~lalit.shukla_logicmonitor.com/opentelemetry-go-contrib.git v0.0.0-20210906044236-4761ee169b1f
)
