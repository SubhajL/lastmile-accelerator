package startup

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadiness_Check_ReturnsReadyWhenAllCriticalPass(t *testing.T) {
	r := NewReadiness(map[string]CheckFunc{
		"vault": func(ctx context.Context) error { return nil },
	}, nil)

	rep, ready := r.Check(context.Background())
	assert.True(t, ready)
	assert.Equal(t, map[string]string{"vault": "ok"}, rep.Checks)
}

func TestReadiness_Check_NotReadyWhenCriticalFails(t *testing.T) {
	boom := errors.New("boom")
	r := NewReadiness(map[string]CheckFunc{
		"postgres": func(ctx context.Context) error { return boom },
	}, nil)

	rep, ready := r.Check(context.Background())
	assert.False(t, ready)
	assert.Contains(t, rep.Checks["postgres"], "error:")
}

func TestReadiness_Check_OptionalDoesNotGateReady(t *testing.T) {
	r := NewReadiness(map[string]CheckFunc{
		"vault": func(ctx context.Context) error { return nil },
	}, map[string]CheckFunc{
		"redis": func(ctx context.Context) error { return errors.New("down") },
	})

	rep, ready := r.Check(context.Background())
	assert.True(t, ready)
	assert.Equal(t, "ok", rep.Checks["vault"])
	assert.Contains(t, rep.Checks["redis"], "error:")
}
