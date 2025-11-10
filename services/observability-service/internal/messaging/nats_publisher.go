package messaging

import (
	"context"
	"encoding/json"
	"fmt"
)

type natsConnLike interface {
	Publish(subject string, data []byte) error
}

type NATSPublisher struct {
	nc natsConnLike
}

func NewNATSPublisher(nc natsConnLike) *NATSPublisher { return &NATSPublisher{nc: nc} }

func (p *NATSPublisher) PublishAlertTriggered(ctx context.Context, payload map[string]interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := p.nc.Publish("observability.alert.triggered", b); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return nil
}
