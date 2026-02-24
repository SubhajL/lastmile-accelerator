package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeOTLPEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "empty", raw: "", want: "", wantErr: false},
		{name: "whitespace", raw: "  ", want: "", wantErr: false},
		{name: "hostport", raw: "localhost:4317", want: "localhost:4317", wantErr: false},
		{name: "http-scheme", raw: "http://localhost:4317", want: "localhost:4317", wantErr: false},
		{name: "https-scheme", raw: "https://localhost:4317", want: "localhost:4317", wantErr: false},
		{name: "trailing-slash", raw: "http://localhost:4317/", want: "localhost:4317", wantErr: false},
		{name: "reject-path-with-scheme", raw: "http://localhost:4317/v1/traces", wantErr: true},
		{name: "reject-bare-path", raw: "/v1/traces", wantErr: true},
		{name: "reject-host-with-path", raw: "localhost:4317/v1/traces", wantErr: true},
		{name: "reject-invalid-url", raw: "http://%zz", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeOTLPEndpoint(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
