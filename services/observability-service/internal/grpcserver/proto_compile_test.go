package grpcserver

import (
    "testing"

    obs "example.com/lma/observability-service/internal/gen/observability/v1/observability"
    "google.golang.org/protobuf/types/known/durationpb"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// TDD: this compile-time test should fail until we generate protobuf stubs.
func TestProto_GeneratedPackage_Compiles(t *testing.T) {
    _ = obs.SLO{}
    _ = obs.SLOStatus{}
    _ = obs.AlertRule{}
    _ = obs.ErrorGroup{}

    if obs.SLOType_AVAILABILITY != 1 || obs.SLOType_ERROR_RATE != 4 {
        t.Fatalf("SLOType enum values mismatch: got AVAIL=%d ERR=%d", obs.SLOType_AVAILABILITY, obs.SLOType_ERROR_RATE)
    }

    // Sanity: timestamp/duration types available for fields
    _ = &timestamppb.Timestamp{}
    _ = &durationpb.Duration{}
}
