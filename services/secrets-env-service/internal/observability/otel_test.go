package observability

import (
    "context"
    "testing"

    "example.com/lma/secrets-env-service/internal/config"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.opentelemetry.io/otel/sdk/resource"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func TestBuildOTLPClientConfig_ParsesEndpointInsecureHeaders(t *testing.T) {
    cfg := config.ObservabilityConfig{
        OTELEndpoint:   "localhost:4317",
        OTELInsecure:   true,
        OTELHeaders:    map[string]string{"api-key": "abc", "x-team": "core"},
        OTELServiceName: "",
    }
    c := buildOTLPClientConfig(cfg)
    assert.Equal(t, "localhost:4317", c.Endpoint)
    assert.True(t, c.Insecure)
    assert.Equal(t, map[string]string{"api-key": "abc", "x-team": "core"}, c.Headers)
}

func TestBuildResource_UsesOverrideNameWhenProvided(t *testing.T) {
    ctx := context.Background()
    r := buildResource(ctx, "secrets-env-service", "svc-override")
    require.IsType(t, &resource.Resource{}, r)
    // find service.name attribute
    var got string
    for _, a := range r.Attributes() {
        if a.Key == semconv.ServiceNameKey {
            got = a.Value.AsString()
            break
        }
    }
    assert.Equal(t, "svc-override", got)
}

func TestInit_SetsTracerProviderAndShutdownNonNil(t *testing.T) {
    ctx := context.Background()
    cfg := config.ObservabilityConfig{
        OTELEndpoint:   "localhost:4317",
        OTELInsecure:   true,
        OTELHeaders:    map[string]string{"x": "y"},
        OTELServiceName: "",
    }
    shutdown, err := Init(ctx, cfg, "secrets-env-service")
    require.NoError(t, err)
    require.NotNil(t, shutdown)
}
