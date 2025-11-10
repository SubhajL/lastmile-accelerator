package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

// Publisher defines a generic event publisher
type Publisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

type natsConn interface {
	Publish(subj string, data []byte) error
	FlushTimeout(timeout time.Duration) error
	Close()
}

// NATSPublisher implements Publisher using NATS
type NATSPublisher struct{ c natsConn }

type envelope struct {
	Topic       string      `json:"topic"`
	OccurredAt  time.Time   `json:"occurred_at"`
	Traceparent string      `json:"traceparent"`
	Payload     interface{} `json:"payload"`
}

// ctx key for traceparent propagation
type ctxKeyTraceparent struct{}

func WithTraceparent(ctx context.Context, tp string) context.Context { return context.WithValue(ctx, ctxKeyTraceparent{}, tp) }
func TraceparentFromContext(ctx context.Context) string {
	if v := ctx.Value(ctxKeyTraceparent{}); v != nil {
		if s, ok := v.(string); ok { return s }
	}
	// generate a minimal W3C traceparent (not cryptographically random here)
	return "00-00000000000000000000000000000001-0000000000000001-01"
}

func NewNATSPublisherFromURL(url string) (*NATSPublisher, error) {
	conn, err := nats.Connect(url)
	if err != nil { return nil, err }
	return &NATSPublisher{c: conn}, nil
}

// NewNATSPublisherWithConn is for tests
func NewNATSPublisherWithConn(c natsConn) *NATSPublisher { return &NATSPublisher{c: c} }

func (p *NATSPublisher) Publish(ctx context.Context, topic string, payload any) error {
	env := envelope{Topic: topic, OccurredAt: time.Now(), Traceparent: TraceparentFromContext(ctx), Payload: payload}
	b, err := json.Marshal(env)
	if err != nil { return err }
	if err := p.c.Publish(topic, b); err != nil { return err }
	_ = p.c.FlushTimeout(2 * time.Second)
	return nil
}
