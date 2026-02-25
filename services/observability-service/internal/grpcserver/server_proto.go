package grpcserver

import (
    "context"
    "time"

    obs "example.com/lma/observability-service/internal/gen/observability/v1/observability"
    "example.com/lma/observability-service/internal/models"
    "example.com/lma/observability-service/internal/services"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type protoServer struct {
    obs.UnimplementedObservabilityServiceServer
    slo     services.SLOService
    alerts  services.AlertService
    errors  services.ErrorService
    queries TraceQuery
}

func NewProtoObservabilityServer(slo services.SLOService, alerts services.AlertService, errs services.ErrorService, q TraceQuery) obs.ObservabilityServiceServer {
    return &protoServer{slo: slo, alerts: alerts, errors: errs, queries: q}
}

// ---- Implemented RPCs ----

func (s *protoServer) GetSLOStatus(ctx context.Context, in *obs.GetSLOStatusRequest) (*obs.GetSLOStatusResponse, error) {
    if in == nil || in.GetId() == "" { return nil, status.Error(codes.InvalidArgument, "id required") }
    st, err := s.slo.GetSLOStatus(ctx, in.GetId())
    if err != nil || st == nil { return nil, status.Error(codes.NotFound, "not found") }
    return &obs.GetSLOStatusResponse{Status: toProtoSLOStatus(st)}, nil
}

func (s *protoServer) SearchTraces(ctx context.Context, in *obs.SearchTracesRequest) (*obs.SearchTracesResponse, error) {
    if in == nil { return nil, status.Error(codes.InvalidArgument, "request required") }
    var start, end time.Time
    if in.GetRange() != nil { start, end = timeRangeFromProto(in.GetRange()) }
    rows, err := s.queries.SearchTraces(ctx, "", "", 0, 0, false, in.GetQ(), start, end, int(in.GetLimit()))
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.TraceSummary, 0, len(rows))
    for _, r := range rows { out = append(out, toProtoTraceSummary(r)) }
    return &obs.SearchTracesResponse{Traces: out}, nil
}

// ---- SLO RPCs ----
func (s *protoServer) CreateSLO(ctx context.Context, in *obs.CreateSLORequest) (*obs.CreateSLOResponse, error) {
    if in == nil || in.GetSlo()==nil { return nil, status.Error(codes.InvalidArgument, "slo required") }
    m, err := fromProtoSLO(in.GetSlo())
    if err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    if err := s.slo.CreateSLO(ctx, m); err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    return &obs.CreateSLOResponse{Slo: toProtoSLO(m)}, nil
}
func (s *protoServer) GetSLO(ctx context.Context, in *obs.GetSLORequest) (*obs.GetSLOResponse, error) {
    if in == nil || in.GetId()=="" { return nil, status.Error(codes.InvalidArgument, "id required") }
    slo, err := s.slo.GetSLO(ctx, in.GetId())
    if err != nil || slo==nil { return nil, status.Error(codes.NotFound, "not found") }
    return &obs.GetSLOResponse{Slo: toProtoSLO(slo)}, nil
}
func (s *protoServer) ListSLOsByProject(ctx context.Context, in *obs.ListSLOsByProjectRequest) (*obs.ListSLOsByProjectResponse, error) {
    if in == nil || in.GetProjectId()=="" { return nil, status.Error(codes.InvalidArgument, "project_id required") }
    rows, err := s.slo.ListProjectSLOs(ctx, in.GetProjectId())
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.SLO, 0, len(rows))
    for _, r := range rows { out = append(out, toProtoSLO(r)) }
    return &obs.ListSLOsByProjectResponse{Slos: out}, nil
}
func (s *protoServer) UpdateSLO(ctx context.Context, in *obs.UpdateSLORequest) (*obs.UpdateSLOResponse, error) {
    if in == nil || in.GetSlo()==nil || in.GetSlo().GetId()=="" { return nil, status.Error(codes.InvalidArgument, "slo with id required") }
    m, err := fromProtoSLO(in.GetSlo())
    if err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    if err := s.slo.UpdateSLO(ctx, m); err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    return &obs.UpdateSLOResponse{Slo: toProtoSLO(m)}, nil
}
func (s *protoServer) DeleteSLO(ctx context.Context, in *obs.DeleteSLORequest) (*obs.DeleteSLOResponse, error) {
    if in == nil || in.GetId()=="" { return nil, status.Error(codes.InvalidArgument, "id required") }
    if err := s.slo.DeleteSLO(ctx, in.GetId()); err != nil { return nil, status.Error(codes.NotFound, err.Error()) }
    return &obs.DeleteSLOResponse{}, nil
}
func (s *protoServer) GetSLOHistory(ctx context.Context, in *obs.GetSLOHistoryRequest) (*obs.GetSLOHistoryResponse, error) {
    if in == nil || in.GetId()=="" { return nil, status.Error(codes.InvalidArgument, "id required") }
    var from, to time.Time
    if in.GetRange()!=nil { from, to = timeRangeFromProto(in.GetRange()) }
    rows, err := s.slo.GetSLOHistory(ctx, in.GetId(), from, to)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.SLOHistoryPoint, 0, len(rows))
    for _, r := range rows { out = append(out, toProtoHistoryPoint(r)) }
    return &obs.GetSLOHistoryResponse{Points: out}, nil
}

// ---- Alerts RPCs ----
func (s *protoServer) CreateAlert(ctx context.Context, in *obs.CreateAlertRequest) (*obs.CreateAlertResponse, error) {
    if s.alerts == nil { return nil, status.Error(codes.FailedPrecondition, "alerts not wired") }
    if in == nil || in.GetRule()==nil { return nil, status.Error(codes.InvalidArgument, "rule required") }
    m, err := fromProtoAlertRule(in.GetRule())
    if err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    if err := s.alerts.CreateAlert(ctx, m); err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    return &obs.CreateAlertResponse{Rule: toProtoAlertRule(m)}, nil
}
func (s *protoServer) GetAlert(ctx context.Context, in *obs.GetAlertRequest) (*obs.GetAlertResponse, error) {
    if s.alerts == nil { return nil, status.Error(codes.FailedPrecondition, "alerts not wired") }
    if in == nil || in.GetId()=="" { return nil, status.Error(codes.InvalidArgument, "id required") }
    r, err := s.alerts.GetAlert(ctx, in.GetId())
    if err != nil || r==nil { return nil, status.Error(codes.NotFound, "not found") }
    return &obs.GetAlertResponse{Rule: toProtoAlertRule(r)}, nil
}
func (s *protoServer) ListAlertsBySLO(ctx context.Context, in *obs.ListAlertsBySLORequest) (*obs.ListAlertsBySLOResponse, error) {
    if s.alerts == nil { return nil, status.Error(codes.FailedPrecondition, "alerts not wired") }
    if in == nil || in.GetSloId()=="" { return nil, status.Error(codes.InvalidArgument, "slo_id required") }
    rows, err := s.alerts.GetSLOAlerts(ctx, in.GetSloId())
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.AlertRule, 0, len(rows))
    for _, r := range rows { out = append(out, toProtoAlertRule(r)) }
    return &obs.ListAlertsBySLOResponse{Rules: out}, nil
}
func (s *protoServer) UpdateAlert(ctx context.Context, in *obs.UpdateAlertRequest) (*obs.UpdateAlertResponse, error) {
    if s.alerts == nil { return nil, status.Error(codes.FailedPrecondition, "alerts not wired") }
    if in == nil || in.GetRule()==nil || in.GetRule().GetId()=="" { return nil, status.Error(codes.InvalidArgument, "rule with id required") }
    m, err := fromProtoAlertRule(in.GetRule())
    if err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    if err := s.alerts.UpdateAlert(ctx, m); err != nil { return nil, status.Error(codes.InvalidArgument, err.Error()) }
    return &obs.UpdateAlertResponse{Rule: toProtoAlertRule(m)}, nil
}
func (s *protoServer) DeleteAlert(ctx context.Context, in *obs.DeleteAlertRequest) (*obs.DeleteAlertResponse, error) {
    if s.alerts == nil { return nil, status.Error(codes.FailedPrecondition, "alerts not wired") }
    if in == nil || in.GetId()=="" { return nil, status.Error(codes.InvalidArgument, "id required") }
    if err := s.alerts.DeleteAlert(ctx, in.GetId()); err != nil { return nil, status.Error(codes.NotFound, err.Error()) }
    return &obs.DeleteAlertResponse{}, nil
}
func (s *protoServer) GetAlertHistory(ctx context.Context, in *obs.GetAlertHistoryRequest) (*obs.GetAlertHistoryResponse, error) {
    if s.alerts == nil { return nil, status.Error(codes.FailedPrecondition, "alerts not wired") }
    if in == nil || in.GetId()=="" { return nil, status.Error(codes.InvalidArgument, "id required") }
    hist, err := s.alerts.GetAlertHistory(ctx, in.GetId(), 50)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.AlertEvent, 0, len(hist))
    for _, h := range hist { out = append(out, toProtoAlertEvent(h)) }
    return &obs.GetAlertHistoryResponse{Events: out}, nil
}

// ---- Error Inbox RPCs ----
func (s *protoServer) IngestError(ctx context.Context, in *obs.IngestErrorRequest) (*obs.IngestErrorResponse, error) {
    if s.errors == nil { return nil, status.Error(codes.FailedPrecondition, "errors not wired") }
    if in == nil || in.GetProjectId()=="" || in.GetEvent()==nil { return nil, status.Error(codes.InvalidArgument, "project_id and event required") }
    ev := in.GetEvent()
    gid, err := s.errors.Ingest(ctx, in.GetProjectId(), ev.GetMessage(), ev.GetStack(), ev.GetMessage(), ev.GetMessage(), nil)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    return &obs.IngestErrorResponse{GroupId: gid, EventId: ""}, nil
}
func (s *protoServer) ListErrorGroups(ctx context.Context, in *obs.ListErrorGroupsRequest) (*obs.ListErrorGroupsResponse, error) {
    if s.errors == nil { return nil, status.Error(codes.FailedPrecondition, "errors not wired") }
    if in == nil || in.GetProjectId()=="" { return nil, status.Error(codes.InvalidArgument, "project_id required") }
    rows, err := s.errors.ListGroups(ctx, in.GetProjectId(), models.ErrorGroupFilter{})
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.ErrorGroup, 0, len(rows))
    for _, g := range rows { out = append(out, toProtoErrorGroup(g)) }
    return &obs.ListErrorGroupsResponse{Groups: out}, nil
}
func (s *protoServer) GetErrorGroup(ctx context.Context, in *obs.GetErrorGroupRequest) (*obs.GetErrorGroupResponse, error) {
    if s.errors == nil { return nil, status.Error(codes.FailedPrecondition, "errors not wired") }
    if in == nil || in.GetGroupId()=="" { return nil, status.Error(codes.InvalidArgument, "group_id required") }
    g, err := s.errors.GetGroup(ctx, in.GetGroupId())
    if err != nil || g==nil { return nil, status.Error(codes.NotFound, "not found") }
    return &obs.GetErrorGroupResponse{Group: toProtoErrorGroup(*g)}, nil
}
func (s *protoServer) ListGroupEvents(ctx context.Context, in *obs.ListGroupEventsRequest) (*obs.ListGroupEventsResponse, error) {
    if s.errors == nil { return nil, status.Error(codes.FailedPrecondition, "errors not wired") }
    if in == nil || in.GetGroupId()=="" { return nil, status.Error(codes.InvalidArgument, "group_id required") }
    rows, err := s.errors.ListGroupEvents(ctx, in.GetGroupId(), int(in.GetLimit()), 0)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.ErrorEvent, 0, len(rows))
    for _, e := range rows { out = append(out, toProtoErrorEvent(e)) }
    return &obs.ListGroupEventsResponse{Events: out}, nil
}
func (s *protoServer) ResolveGroup(ctx context.Context, in *obs.ResolveGroupRequest) (*obs.ResolveGroupResponse, error) {
    if s.errors == nil { return nil, status.Error(codes.FailedPrecondition, "errors not wired") }
    if in == nil || in.GetGroupId()=="" { return nil, status.Error(codes.InvalidArgument, "group_id required") }
    if err := s.errors.ResolveGroup(ctx, in.GetGroupId()); err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    g, _ := s.errors.GetGroup(ctx, in.GetGroupId())
    var pg *obs.ErrorGroup
    if g != nil { pg = toProtoErrorGroup(*g) }
    return &obs.ResolveGroupResponse{Group: pg}, nil
}

func (s *protoServer) GetTrace(ctx context.Context, in *obs.GetTraceRequest) (*obs.GetTraceResponse, error) {
    if in == nil || in.GetTraceId()=="" { return nil, status.Error(codes.InvalidArgument, "trace_id required") }
    d, err := s.queries.GetTrace(ctx, in.GetTraceId())
    if err != nil || d==nil { return nil, status.Error(codes.NotFound, "not found") }
    return &obs.GetTraceResponse{Trace: toProtoTrace(*d)}, nil
}
func (s *protoServer) SearchLogs(ctx context.Context, in *obs.SearchLogsRequest) (*obs.SearchLogsResponse, error) {
    if in == nil || in.GetQ()=="" { return nil, status.Error(codes.InvalidArgument, "q required") }
    var start, end time.Time
    if in.GetRange()!=nil { start, end = timeRangeFromProto(in.GetRange()) }
    rows, err := s.queries.SearchLogs(ctx, in.GetQ(), start, end, int(in.GetLimit()), "desc")
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    out := make([]*obs.LogEntry, 0, len(rows))
    for _, e := range rows { out = append(out, toProtoLogEntry(e)) }
    return &obs.SearchLogsResponse{Logs: out}, nil
}
func (s *protoServer) Golden(ctx context.Context, in *obs.GoldenRequest) (*obs.GoldenResponse, error) {
    if in == nil || in.GetProjectId()=="" { return nil, status.Error(codes.InvalidArgument, "project_id required") }
    g, err := s.queries.Golden(ctx, in.GetProjectId(), 5*time.Minute)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    return toProtoGolden(g), nil
}
