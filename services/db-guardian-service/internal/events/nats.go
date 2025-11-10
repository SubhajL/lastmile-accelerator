package events

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

func NewNATSClient(url string) (*nats.Conn, error) {
	if url == "" {
		return nil, fmt.Errorf("NATS URL is required")
	}

	nc, err := nats.Connect(url,
		nats.Timeout(5*time.Second),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(3),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return nc, nil
}

func PublishEvent(ctx context.Context, nc *nats.Conn, subject string, data []byte) error {
	if nc == nil {
		return fmt.Errorf("NATS connection is nil")
	}

	// Extract trace context if available
	spanCtx := trace.SpanContextFromContext(ctx)
	var msg *nats.Msg
	
	if spanCtx.IsValid() {
		msg = &nats.Msg{
			Subject: subject,
			Data:    data,
			Header:  nats.Header{},
		}
		// Add traceparent header for distributed tracing
		msg.Header.Set("traceparent", formatTraceparent(spanCtx))
	} else {
		msg = &nats.Msg{
			Subject: subject,
			Data:    data,
		}
	}

	if err := nc.PublishMsg(msg); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func formatTraceparent(spanCtx trace.SpanContext) string {
	// Format: version-traceID-spanID-flags
	return fmt.Sprintf("00-%s-%s-%02x",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		spanCtx.TraceFlags(),
	)
}
