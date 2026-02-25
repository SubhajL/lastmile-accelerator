package startup

import (
	"context"
	"fmt"
)

type CheckFunc func(ctx context.Context) error

type Readiness struct {
	critical map[string]CheckFunc
	optional map[string]CheckFunc
}

type Report struct {
	Checks map[string]string `json:"checks"`
}

func NewReadiness(critical map[string]CheckFunc, optional map[string]CheckFunc) Readiness {
	if critical == nil {
		critical = map[string]CheckFunc{}
	}
	if optional == nil {
		optional = map[string]CheckFunc{}
	}
	return Readiness{critical: critical, optional: optional}
}

func (r Readiness) Check(ctx context.Context) (Report, bool) {
	report := Report{Checks: map[string]string{}}
	ready := true

	for name, fn := range r.critical {
		if err := fn(ctx); err != nil {
			report.Checks[name] = fmt.Sprintf("error: %v", err)
			ready = false
		} else {
			report.Checks[name] = "ok"
		}
	}
	for name, fn := range r.optional {
		if err := fn(ctx); err != nil {
			report.Checks[name] = fmt.Sprintf("error: %v", err)
		} else {
			report.Checks[name] = "ok"
		}
	}

	return report, ready
}
