package grpcserver

import (
    "encoding/json"
    "fmt"
    "time"

    obs "example.com/lma/observability-service/internal/gen/observability/v1/observability"
    "example.com/lma/observability-service/internal/logs"
    "example.com/lma/observability-service/internal/models"
    "example.com/lma/observability-service/internal/services"
    "example.com/lma/observability-service/internal/traces"
    "google.golang.org/protobuf/types/known/durationpb"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoSLOStatus(st *models.SLOStatus) *obs.SLOStatus {
    if st == nil { return nil }
    return &obs.SLOStatus{
        SloId:           st.SLOID,
        Compliance:      st.Compliance,
        BurnRate:        st.BurnRate,
        RemainingBudget: st.RemainingBudget,
        LastCalculated:  timestamppb.New(st.LastCalculated),
    }
}

func timeRangeFromProto(r *obs.TimeRange) (time.Time, time.Time) {
    if r == nil { return time.Time{}, time.Time{} }
    var from, to time.Time
    if r.From != nil { from = r.From.AsTime() }
    if r.To != nil { to = r.To.AsTime() }
    return from, to
}

func toProtoTraceSummary(s traces.TraceSummary) *obs.TraceSummary {
    return &obs.TraceSummary{
        TraceId:     s.TraceID,
        ServiceName: s.Service,
        DurationMs:  s.DurationMs,
        Attrs:       map[string]string{"operation": s.Operation},
    }
}

// ---- SLO mappers ----
func toProtoSLO(m *models.SLO) *obs.SLO {
    if m == nil { return nil }
    return &obs.SLO{
        Id: m.ID,
        ProjectId: m.ProjectID,
        ServiceName: m.ServiceName,
        Type: toProtoSLOType(m.Type),
        Target: m.Target,
        Window: durationpb.New(m.Window),
        Query: m.Query,
        Status: m.Status,
        CreatedAt: timestamppb.New(m.CreatedAt),
        UpdatedAt: timestamppb.New(m.UpdatedAt),
    }
}

func fromProtoSLO(p *obs.SLO) (*models.SLO, error) {
    if p == nil { return nil, fmt.Errorf("slo required") }
    return &models.SLO{
        ID: p.GetId(),
        ProjectID: p.GetProjectId(),
        ServiceName: p.GetServiceName(),
        Type: fromProtoSLOType(p.GetType()),
        Target: p.GetTarget(),
        Window: p.GetWindow().AsDuration(),
        Query: p.GetQuery(),
        Status: p.GetStatus(),
        CreatedAt: tsOrZero(p.GetCreatedAt()),
        UpdatedAt: tsOrZero(p.GetUpdatedAt()),
    }, nil
}

func toProtoHistoryPoint(h *models.SLOHistory) *obs.SLOHistoryPoint {
    if h == nil { return nil }
    return &obs.SLOHistoryPoint{SloId: h.SLOID, Timestamp: timestamppb.New(h.Timestamp), Compliance: h.Compliance, BurnRate: h.BurnRate}
}

func toProtoSLOType(t models.SLOType) obs.SLOType {
    switch t {
    case models.SLOTypeAvailability: return obs.SLOType_AVAILABILITY
    case models.SLOTypeLatency: return obs.SLOType_LATENCY
    case models.SLOTypeThroughput: return obs.SLOType_THROUGHPUT
    case models.SLOTypeErrorRate: return obs.SLOType_ERROR_RATE
    default: return obs.SLOType_SLO_TYPE_UNSPECIFIED
    }
}

func fromProtoSLOType(t obs.SLOType) models.SLOType {
    switch t {
    case obs.SLOType_AVAILABILITY: return models.SLOTypeAvailability
    case obs.SLOType_LATENCY: return models.SLOTypeLatency
    case obs.SLOType_THROUGHPUT: return models.SLOTypeThroughput
    case obs.SLOType_ERROR_RATE: return models.SLOTypeErrorRate
    default: return ""
    }
}

func tsOrZero(ts *timestamppb.Timestamp) time.Time {
    if ts == nil { return time.Time{} }
    return ts.AsTime()
}

// ---- Alerts mappers ----
func toProtoAlertRule(a *models.AlertRule) *obs.AlertRule {
    if a == nil { return nil }
    return &obs.AlertRule{
        Id: a.ID,
        SloId: a.SLOID,
        Threshold: a.Threshold,
        Channels: toProtoAlertChannels(a.Channels),
        CreatedAt: timestamppb.New(a.CreatedAt),
        UpdatedAt: timestamppb.New(a.UpdatedAt),
    }
}

func fromProtoAlertRule(p *obs.AlertRule) (*models.AlertRule, error) {
    if p == nil { return nil, fmt.Errorf("rule required") }
    return &models.AlertRule{
        ID: p.GetId(),
        SLOID: p.GetSloId(),
        Threshold: p.GetThreshold(),
        Channels: fromProtoAlertChannels(p.GetChannels()),
        Enabled: true,
        CreatedAt: tsOrZero(p.GetCreatedAt()),
        UpdatedAt: tsOrZero(p.GetUpdatedAt()),
    }, nil
}

func toProtoAlertEvent(h *models.AlertHistory) *obs.AlertEvent {
    if h == nil { return nil }
    status := "ok"
    if h.Notified { status = "triggered" }
    return &obs.AlertEvent{Id: h.ID, AlertId: h.AlertRuleID, Status: status, Timestamp: timestamppb.New(h.Timestamp), Message: ""}
}

func toProtoAlertChannels(ch []models.AlertChannel) []obs.AlertChannel {
    out := make([]obs.AlertChannel, 0, len(ch))
    for _, c := range ch {
        switch c {
        case models.AlertChannelEmail: out = append(out, obs.AlertChannel_EMAIL)
        case models.AlertChannelSlack: out = append(out, obs.AlertChannel_SLACK)
        case models.AlertChannelPagerDuty: out = append(out, obs.AlertChannel_PAGERDUTY)
        case models.AlertChannelWebhook: // no direct mapping; skip or treat as SLACK
            out = append(out, obs.AlertChannel_SLACK)
        default:
            out = append(out, obs.AlertChannel_ALERT_CHANNEL_UNSPECIFIED)
        }
    }
    return out
}

func fromProtoAlertChannels(ch []obs.AlertChannel) []models.AlertChannel {
    out := make([]models.AlertChannel, 0, len(ch))
    for _, c := range ch {
        switch c {
        case obs.AlertChannel_EMAIL: out = append(out, models.AlertChannelEmail)
        case obs.AlertChannel_SLACK: out = append(out, models.AlertChannelSlack)
        case obs.AlertChannel_PAGERDUTY: out = append(out, models.AlertChannelPagerDuty)
        default:
            // ignore unspecified
        }
    }
    if len(out) == 0 { out = []models.AlertChannel{models.AlertChannelEmail} }
    return out
}

// ---- Error inbox mappers ----
func toProtoErrorGroup(g models.ErrorGroup) *obs.ErrorGroup {
    // last seen string -> timestamp
    var ts time.Time
    if g.LastSeen != "" { if t, err := time.Parse(time.RFC3339, g.LastSeen); err == nil { ts = t } }
    resolved := g.Status == "resolved"
    return &obs.ErrorGroup{Id: g.ID, ProjectId: g.ProjectID, Fingerprint: g.Fingerprint, Count: g.Occurrences, LastSeen: timestamppb.New(ts), Resolved: resolved}
}

func toProtoErrorEvent(e models.ErrorEvent) *obs.ErrorEvent {
    var ts time.Time
    if e.TS != "" { if t, err := time.Parse(time.RFC3339, e.TS); err == nil { ts = t } }
    attrs := make(map[string]string, len(e.Metadata))
    for k, v := range e.Metadata { attrs[k] = fmt.Sprint(v) }
    return &obs.ErrorEvent{Id: e.ID, GroupId: e.GroupID, Message: e.Message, Stack: e.Stack, Timestamp: timestamppb.New(ts), Attrs: attrs}
}

// ---- Traces/Logs/Dashboard mappers ----
func toProtoTrace(d traces.TraceDetail) *obs.Trace {
    // encode spans as JSON array payload
    b, _ := json.Marshal(d.Spans)
    return &obs.Trace{TraceId: d.TraceID, Payload: b}
}

func toProtoLogEntry(e logs.LogEntry) *obs.LogEntry {
    return &obs.LogEntry{Ts: timestamppb.New(e.Timestamp), Message: e.Line, Labels: e.Labels}
}

func toProtoGolden(g *services.GoldenSignals) *obs.GoldenResponse {
    if g == nil { return &obs.GoldenResponse{} }
    return &obs.GoldenResponse{Dashboard: &obs.GoldenDashboard{Metrics: map[string]float64{"request_rate": g.RequestRate, "error_rate": g.ErrorRate, "p95_latency": g.P95Latency}}}
}
