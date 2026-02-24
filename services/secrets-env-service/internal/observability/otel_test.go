package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeOTLPEndpoint(t *testing.T) {
	got, err := normalizeOTLPEndpoint("localhost:4317")
	require.NoError(t, err)
	assert.Equal(t, "localhost:4317", got)
}

func TestNormalizeOTLPEndpoint_StripsScheme(t *testing.T) {
	got, err := normalizeOTLPEndpoint("http://localhost:4317")
	require.NoError(t, err)
	assert.Equal(t, "localhost:4317", got)
}

func TestNormalizeOTLPEndpoint_RejectsPath(t *testing.T) {
	_, err := normalizeOTLPEndpoint("http://localhost:4317/v1/traces")
	assert.Error(t, err)
}
