package grpcserver

import (
    "context"
    "strings"

    "example.com/lma/observability-service/internal/middleware"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

// UnaryAuthInterceptor enforces JWT auth and scope checks on incoming RPCs.
func UnaryAuthInterceptor(verifier middleware.Verifier) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok { return nil, status.Error(codes.Unauthenticated, "missing metadata") }
        auths := md.Get("authorization")
        if len(auths) == 0 { return nil, status.Error(codes.Unauthenticated, "missing authorization") }
        auth := auths[0]
        if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
            return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
        }
        tok := strings.TrimSpace(auth[len("bearer "):])
        claims, err := verifier.Verify(ctx, tok)
        if err != nil { return nil, status.Error(codes.Unauthenticated, "invalid token") }
        for _, need := range requiredScopesFor(info.FullMethod) {
            if !contains(claims.Scopes, need) {
                return nil, status.Error(codes.PermissionDenied, "forbidden")
            }
        }
        return handler(ctx, req)
    }
}

func contains(ss []string, s string) bool { for _, x := range ss { if x == s { return true } } ; return false }

// requiredScopesFor maps methods to required scopes; it matches by suffix so it
// works for both legacy method names and the generated proto service.
func requiredScopesFor(fullMethod string) []string {
    switch {
    // read-only
    case strings.HasSuffix(fullMethod, "/GetSLOStatus"),
        strings.HasSuffix(fullMethod, "/SearchTraces"),
        strings.HasSuffix(fullMethod, "/GetTrace"),
        strings.HasSuffix(fullMethod, "/SearchLogs"),
        strings.HasSuffix(fullMethod, "/Golden"),
        strings.HasSuffix(fullMethod, "/GetSLO"),
        strings.HasSuffix(fullMethod, "/ListSLOsByProject"),
        strings.HasSuffix(fullMethod, "/GetAlert"),
        strings.HasSuffix(fullMethod, "/ListAlertsBySLO"),
        strings.HasSuffix(fullMethod, "/GetAlertHistory"),
        strings.HasSuffix(fullMethod, "/ListErrorGroups"),
        strings.HasSuffix(fullMethod, "/GetErrorGroup"),
        strings.HasSuffix(fullMethod, "/ListGroupEvents"):
        return []string{"observability:read"}
    // write
    case strings.HasSuffix(fullMethod, "/CreateSLO"),
        strings.HasSuffix(fullMethod, "/UpdateSLO"),
        strings.HasSuffix(fullMethod, "/DeleteSLO"),
        strings.HasSuffix(fullMethod, "/CreateAlert"),
        strings.HasSuffix(fullMethod, "/UpdateAlert"),
        strings.HasSuffix(fullMethod, "/DeleteAlert"):
        return []string{"observability:write"}
    // ingest
    case strings.HasSuffix(fullMethod, "/IngestError"):
        return []string{"observability:ingest"}
    default:
        return []string{"observability:read"}
    }
}
