package grpcserver

import (
	"context"
	"encoding/json"
	"time"

	"example.com/lma/observability-service/internal/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Requests/Responses (JSON-encoded for tests; mirrors eventual proto fields)
type GetSLOStatusRequest struct { SloId string `json:"slo_id"` }
type GetSLOStatusResponse struct {
	SloId           string  `json:"slo_id"`
	Compliance      float64 `json:"compliance"`
	BurnRate        float64 `json:"burn_rate"`
	RemainingBudget float64 `json:"remaining_budget"`
}

type TraceSummary struct {
	TraceID    string `json:"trace_id"`
	Service    string `json:"service"`
	Operation  string `json:"operation"`
	DurationMs int64  `json:"duration_ms"`
	Error      bool   `json:"error"`
}

type SearchTracesRequest struct {
	Service    string `json:"service"`
	Operation  string `json:"operation"`
	ErrorOnly  bool   `json:"error_only"`
	Limit      int    `json:"limit"`
	Query      string `json:"query"`
	StartUnixN int64  `json:"start_unix_nano"`
	EndUnixN   int64  `json:"end_unix_nano"`
}

type SearchTracesResponse struct { Traces []TraceSummary `json:"traces"` }

type GetTraceRequest struct { TraceId string `json:"trace_id"` }

type GetTraceResponse struct {
	TraceId string            `json:"trace_id"`
	Spans   []json.RawMessage `json:"spans"`
}

type LogEntry struct {
	Timestamp int64             `json:"ts_unix_nano"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type SearchLogsRequest struct {
	Q         string `json:"q"`
	Limit     int    `json:"limit"`
	Direction string `json:"direction"`
	StartUnixN int64 `json:"start_unix_nano"`
	EndUnixN   int64 `json:"end_unix_nano"`
}

type SearchLogsResponse struct { Entries []LogEntry `json:"entries"` }

type GoldenRequest struct { Service string `json:"service"`; Window string `json:"window"` }

type GoldenResponse struct {
	RequestRate float64 `json:"request_rate"`
	ErrorRate   float64 `json:"error_rate"`
	P95Latency  float64 `json:"p95_latency"`
}

type ObservabilityServer interface {
	GetSLOStatus(ctx context.Context, in *GetSLOStatusRequest) (*GetSLOStatusResponse, error)
	SearchTraces(ctx context.Context, in *SearchTracesRequest) (*SearchTracesResponse, error)
	GetTrace(ctx context.Context, in *GetTraceRequest) (*GetTraceResponse, error)
	SearchLogs(ctx context.Context, in *SearchLogsRequest) (*SearchLogsResponse, error)
	Golden(ctx context.Context, in *GoldenRequest) (*GoldenResponse, error)
}

type observabilityServer struct {
	sloSvc    services.SLOService
	queries   TraceQuery
}

func NewObservabilityServer(slo services.SLOService, q TraceQuery) ObservabilityServer {
	return &observabilityServer{sloSvc: slo, queries: q}
}

func (s *observabilityServer) GetSLOStatus(ctx context.Context, in *GetSLOStatusRequest) (*GetSLOStatusResponse, error) {
	if in == nil || in.SloId == "" { return nil, status.Error(codes.InvalidArgument, "slo_id required") }
	st, err := s.sloSvc.GetSLOStatus(ctx, in.SloId)
	if err != nil || st == nil { return nil, status.Error(codes.NotFound, "not found") }
	return &GetSLOStatusResponse{SloId: st.SLOID, Compliance: st.Compliance, BurnRate: st.BurnRate, RemainingBudget: st.RemainingBudget}, nil
}

func (s *observabilityServer) SearchTraces(ctx context.Context, in *SearchTracesRequest) (*SearchTracesResponse, error) {
	if in == nil { return nil, status.Error(codes.InvalidArgument, "request required") }
	var start, end time.Time
	if in.StartUnixN != 0 { start = time.Unix(0, in.StartUnixN) }
	if in.EndUnixN != 0 { end = time.Unix(0, in.EndUnixN) }
	res, err := s.queries.SearchTraces(ctx, in.Service, in.Operation, 0, 0, in.ErrorOnly, in.Query, start, end, in.Limit)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	out := make([]TraceSummary, 0, len(res))
	for _, t := range res {
		out = append(out, TraceSummary{TraceID: t.TraceID, Service: t.Service, Operation: t.Operation, DurationMs: t.DurationMs, Error: t.Error})
	}
return &SearchTracesResponse{Traces: out}, nil
}

func (s *observabilityServer) GetTrace(ctx context.Context, in *GetTraceRequest) (*GetTraceResponse, error) {
	if in == nil || in.TraceId == "" { return nil, status.Error(codes.InvalidArgument, "trace_id required") }
	d, err := s.queries.GetTrace(ctx, in.TraceId)
	if err != nil || d == nil { return nil, status.Error(codes.NotFound, "not found") }
	return &GetTraceResponse{TraceId: d.TraceID, Spans: d.Spans}, nil
}

func (s *observabilityServer) SearchLogs(ctx context.Context, in *SearchLogsRequest) (*SearchLogsResponse, error) {
	if in == nil || in.Q == "" { return nil, status.Error(codes.InvalidArgument, "q required") }
	var start, end time.Time
	if in.StartUnixN != 0 { start = time.Unix(0, in.StartUnixN) }
	if in.EndUnixN != 0 { end = time.Unix(0, in.EndUnixN) }
	rows, err := s.queries.SearchLogs(ctx, in.Q, start, end, in.Limit, in.Direction)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	out := make([]LogEntry, 0, len(rows))
	for _, e := range rows { out = append(out, LogEntry{Timestamp: e.Timestamp.UnixNano(), Line: e.Line, Labels: e.Labels}) }
	return &SearchLogsResponse{Entries: out}, nil
}

func (s *observabilityServer) Golden(ctx context.Context, in *GoldenRequest) (*GoldenResponse, error) {
	if in == nil || in.Service == "" { return nil, status.Error(codes.InvalidArgument, "service required") }
	w := in.Window
	if w == "" { w = "5m" }
	dur, err := time.ParseDuration(w)
	if err != nil { return nil, status.Error(codes.InvalidArgument, "invalid window") }
	g, err := s.queries.Golden(ctx, in.Service, dur)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	return &GoldenResponse{RequestRate: g.RequestRate, ErrorRate: g.ErrorRate, P95Latency: g.P95Latency}, nil
}

// gRPC plumbing using a ServiceDesc for JSON-coded messages.
const serviceName = "observability.v1.Observability"

var _Observability_serviceDesc = grpc.ServiceDesc{
	ServiceName: serviceName,
	HandlerType: (*ObservabilityServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetSLOStatus", Handler: _GetSLOStatus_Handler},
		{MethodName: "SearchTraces", Handler: _SearchTraces_Handler},
		{MethodName: "GetTrace", Handler: _GetTrace_Handler},
		{MethodName: "SearchLogs", Handler: _SearchLogs_Handler},
		{MethodName: "Golden", Handler: _Golden_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "observability.json",
}

func RegisterObservabilityServer(s *grpc.Server, srv ObservabilityServer) {
	s.RegisterService(&_Observability_serviceDesc, srv)
}

func _GetSLOStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSLOStatusRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil {
		return srv.(ObservabilityServer).GetSLOStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/"+serviceName+"/GetSLOStatus"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(ObservabilityServer).GetSLOStatus(ctx, req.(*GetSLOStatusRequest)) }
	return interceptor(ctx, in, info, h)
}

func _SearchTraces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchTracesRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil {
		return srv.(ObservabilityServer).SearchTraces(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/"+serviceName+"/SearchTraces"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(ObservabilityServer).SearchTraces(ctx, req.(*SearchTracesRequest)) }
	return interceptor(ctx, in, info, h)
}

func _GetTrace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTraceRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(ObservabilityServer).GetTrace(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/"+serviceName+"/GetTrace"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(ObservabilityServer).GetTrace(ctx, req.(*GetTraceRequest)) }
	return interceptor(ctx, in, info, h)
}

func _SearchLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchLogsRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(ObservabilityServer).SearchLogs(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/"+serviceName+"/SearchLogs"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(ObservabilityServer).SearchLogs(ctx, req.(*SearchLogsRequest)) }
	return interceptor(ctx, in, info, h)
}

func _Golden_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GoldenRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(ObservabilityServer).Golden(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/"+serviceName+"/Golden"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(ObservabilityServer).Golden(ctx, req.(*GoldenRequest)) }
	return interceptor(ctx, in, info, h)
}

