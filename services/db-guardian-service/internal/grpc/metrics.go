package grpcserver

import (
    "context"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
)

var (
    grpcRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total gRPC unary requests by method and code",
        },
        []string{"method", "code"},
    )
    grpcRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "Duration of gRPC unary requests in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "code"},
    )
)

func init() {
    prometheus.MustRegister(grpcRequestsTotal, grpcRequestDuration)
}

// MetricsUnaryInterceptor records per-request counters and latency histograms.
func MetricsUnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        code := status.Code(err)
        labels := []string{info.FullMethod, code.String()}
        grpcRequestsTotal.WithLabelValues(labels...).Inc()
        grpcRequestDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
        return resp, err
    }
}
