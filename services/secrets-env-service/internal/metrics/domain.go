package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

// DomainMetrics holds low-cardinality counters for domain operations.
type DomainMetrics struct {
    secrets *prometheus.CounterVec
    parity  *prometheus.CounterVec
    leaks   *prometheus.CounterVec
}

// Default is the process-wide metrics instance used by services.
var Default *DomainMetrics

// Init creates counters and registers them with the provided Registerer.
// Safe to call multiple times; reuses collectors if already registered.
func Init(reg prometheus.Registerer) *DomainMetrics {
    m := &DomainMetrics{
        secrets: prometheus.NewCounterVec(
            prometheus.CounterOpts{Name: "secrets_operations_total", Help: "Count of secret operations by op and outcome."},
            []string{"op", "outcome"},
        ),
        parity: prometheus.NewCounterVec(
            prometheus.CounterOpts{Name: "parity_operations_total", Help: "Count of parity operations by op and outcome."},
            []string{"op", "outcome"},
        ),
        leaks: prometheus.NewCounterVec(
            prometheus.CounterOpts{Name: "leak_operations_total", Help: "Count of leak-scan operations by op and outcome."},
            []string{"op", "outcome"},
        ),
    }
    // Register with reuse on AlreadyRegistered
    m.secrets = registerOrReuseCounterVec(reg, m.secrets)
    m.parity = registerOrReuseCounterVec(reg, m.parity)
    m.leaks = registerOrReuseCounterVec(reg, m.leaks)
    Default = m
    return m
}

func registerOrReuseCounterVec(reg prometheus.Registerer, cv *prometheus.CounterVec) *prometheus.CounterVec {
    if err := reg.Register(cv); err != nil {
        if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
            if existing, ok := are.ExistingCollector.(*prometheus.CounterVec); ok { return existing }
        }
    }
    return cv
}

// ObserveSecrets increments secrets counter for op/outcome.
func (m *DomainMetrics) ObserveSecrets(op, outcome string) { if m != nil { m.secrets.WithLabelValues(op, outcome).Inc() } }
// ObserveParity increments parity counter for op/outcome.
func (m *DomainMetrics) ObserveParity(op, outcome string)  { if m != nil { m.parity.WithLabelValues(op, outcome).Inc() } }
// ObserveLeaks increments leak counter for op/outcome.
func (m *DomainMetrics) ObserveLeaks(op, outcome string)   { if m != nil { m.leaks.WithLabelValues(op, outcome).Inc() } }
