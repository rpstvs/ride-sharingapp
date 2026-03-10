package tracing

import (
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

func WithTracingInterceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(newServerHandler()),
	}
}

func DialOptionsWithTracing() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithStatsHandler(newClientHandler()),
	}
}

func newServerHandler() stats.Handler {
	return otelgrpc.NewServerHandler(otelgrpc.WithTraceProvider(otel.GetTracerProvider()))
}

func newClientHanlder() stats.Handler {
	return otelgrpc.NewClientHandler(otelgrpc.WithTraceProvider(otel.GetTracerProvider()))
}
