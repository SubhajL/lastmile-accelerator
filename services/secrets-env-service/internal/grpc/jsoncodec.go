package grpcapi

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// jsonCodec provides a simple JSON codec for gRPC used only in tests/internal communication.
// This avoids requiring compiled protobufs for our internal tests.
type jsonCodec struct{}

func (jsonCodec) Name() string { return "json" }
func (jsonCodec) Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

func init() {
	encoding.RegisterCodec(jsonCodec{})
}