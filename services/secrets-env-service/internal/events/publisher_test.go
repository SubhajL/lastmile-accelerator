package events

import (
	"bytes"
	"context"
	"testing"
	"time"
)

type fakeNatsConn struct{ published map[string][]byte }

func (f *fakeNatsConn) Publish(subj string, data []byte) error {
	if f.published == nil { f.published = map[string][]byte{} }
	f.published[subj] = data
	return nil
}
func (f *fakeNatsConn) FlushTimeout(timeout time.Duration) error { return nil }
func (f *fakeNatsConn) Close() {}

func TestNATSPublisher_Publish_JSON(t *testing.T) {
	conn := &fakeNatsConn{}
	pub := NewNATSPublisherWithConn(conn)
	topic := "test.topic"
	payload := map[string]any{"x": 1}
	ctx := WithTraceparent(context.Background(), "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	if err := pub.Publish(ctx, topic, payload); err != nil { t.Fatalf("publish error: %v", err) }
	if b, ok := conn.published[topic]; !ok {
		t.Fatalf("expected publish to %s", topic)
	} else {
		if !bytes.Contains(b, []byte("\"traceparent\"")) { t.Fatalf("expected traceparent in envelope: %s", string(b)) }
	}
}
