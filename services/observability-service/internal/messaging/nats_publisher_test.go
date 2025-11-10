package messaging

import (
	"context"
	"errors"
	"testing"
)

type fakeNats struct {
	lastSubject string
	data        []byte
	err         error
}

func (f *fakeNats) Publish(subject string, data []byte) error {
	f.lastSubject = subject
	f.data = data
	return f.err
}

func TestNATSPublisher_Publish_Success(t *testing.T) {
	f := &fakeNats{}
	pub := NewNATSPublisher(f)
	payload := map[string]interface{}{"x": 1}
	if err := pub.PublishAlertTriggered(context.Background(), payload); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if f.lastSubject != "observability.alert.triggered" {
		t.Fatalf("wrong subject: %s", f.lastSubject)
	}
	if len(f.data) == 0 {
		t.Fatal("expected data")
	}
}

func TestNATSPublisher_Publish_ErrorBubbles(t *testing.T) {
	f := &fakeNats{err: errors.New("boom")}
	pub := NewNATSPublisher(f)
	if err := pub.PublishAlertTriggered(context.Background(), map[string]interface{}{}); err == nil {
		t.Fatal("expected error")
	}
}
