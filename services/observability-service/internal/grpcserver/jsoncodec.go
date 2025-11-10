package grpcserver

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

type jsonCodec struct{}

func (jsonCodec) Name() string { return "json" }
func (jsonCodec) Marshal(v interface{}) ([]byte, error)   { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

// RegisterJSONCodec registers the JSON codec for gRPC so client/server can ForceCodec(json).
func RegisterJSONCodec() {
	encoding.RegisterCodec(jsonCodec{})
}
