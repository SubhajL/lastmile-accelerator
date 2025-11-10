// Code generated manually for testability without protoc. DO NOT EDIT.
package secretsenvv1

import (
	"context"

	"google.golang.org/grpc"
)

// Messages (struct shapes matching .proto).

type SecretMeta struct {
	Id          string `json:"id,omitempty"`
	Key         string `json:"key,omitempty"`
	Environment string `json:"environment,omitempty"`
	// created_at omitted due to manual wiring – carried via domain mapping elsewhere if needed for tests
}

type GetSecretRequest struct {
	TenantId    string `json:"tenant_id,omitempty"`
	ProjectId   string `json:"project_id,omitempty"`
	Key         string `json:"key,omitempty"`
	Environment string `json:"environment,omitempty"`
}

type GetSecretResponse struct {
	Meta  *SecretMeta        `json:"meta,omitempty"`
	Value map[string]any     `json:"value,omitempty"`
}

type ListSecretsRequest struct {
	ProjectId   string `json:"project_id,omitempty"`
	Environment string `json:"environment,omitempty"`
	PageSize    int32  `json:"page_size,omitempty"`
	PageToken   string `json:"page_token,omitempty"`
}

type ListSecretsResponse struct {
	Items         []*SecretMeta `json:"items,omitempty"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}

type CheckEnvParityRequest struct {
	ProjectId  string `json:"project_id,omitempty"`
	BaseEnv    string `json:"base_env,omitempty"`
	CompareEnv string `json:"compare_env,omitempty"`
}

type CheckEnvParityResponse struct {
	ProjectId     string   `json:"project_id,omitempty"`
	MissingKeys   []string `json:"missing_keys,omitempty"`
	ExtraKeys     []string `json:"extra_keys,omitempty"`
	HasDrift      bool     `json:"has_drift,omitempty"`
}

// Client API

type SecretsEnvServiceClient interface {
	GetSecret(ctx context.Context, in *GetSecretRequest, opts ...grpc.CallOption) (*GetSecretResponse, error)
	ListSecrets(ctx context.Context, in *ListSecretsRequest, opts ...grpc.CallOption) (*ListSecretsResponse, error)
	CheckEnvParity(ctx context.Context, in *CheckEnvParityRequest, opts ...grpc.CallOption) (*CheckEnvParityResponse, error)
}

type secretsEnvServiceClient struct{ cc *grpc.ClientConn }

func NewSecretsEnvServiceClient(cc *grpc.ClientConn) SecretsEnvServiceClient { return &secretsEnvServiceClient{cc} }

func (c *secretsEnvServiceClient) GetSecret(ctx context.Context, in *GetSecretRequest, opts ...grpc.CallOption) (*GetSecretResponse, error) {
	out := new(GetSecretResponse)
	err := c.cc.Invoke(ctx, "/lma.secretsenv.v1.SecretsEnvService/GetSecret", in, out, opts...)
	if err != nil { return nil, err }
	return out, nil
}
func (c *secretsEnvServiceClient) ListSecrets(ctx context.Context, in *ListSecretsRequest, opts ...grpc.CallOption) (*ListSecretsResponse, error) {
	out := new(ListSecretsResponse)
	err := c.cc.Invoke(ctx, "/lma.secretsenv.v1.SecretsEnvService/ListSecrets", in, out, opts...)
	if err != nil { return nil, err }
	return out, nil
}
func (c *secretsEnvServiceClient) CheckEnvParity(ctx context.Context, in *CheckEnvParityRequest, opts ...grpc.CallOption) (*CheckEnvParityResponse, error) {
	out := new(CheckEnvParityResponse)
	err := c.cc.Invoke(ctx, "/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity", in, out, opts...)
	if err != nil { return nil, err }
	return out, nil
}

// Server API

type SecretsEnvServiceServer interface {
	GetSecret(context.Context, *GetSecretRequest) (*GetSecretResponse, error)
	ListSecrets(context.Context, *ListSecretsRequest) (*ListSecretsResponse, error)
	CheckEnvParity(context.Context, *CheckEnvParityRequest) (*CheckEnvParityResponse, error)
}

type UnimplementedSecretsEnvServiceServer struct{}

func RegisterSecretsEnvServiceServer(s *grpc.Server, srv SecretsEnvServiceServer) {
	s.RegisterService(&_SecretsEnvService_serviceDesc, srv)
}

func _SecretsEnvService_GetSecret_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSecretRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(SecretsEnvServiceServer).GetSecret(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/lma.secretsenv.v1.SecretsEnvService/GetSecret"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(SecretsEnvServiceServer).GetSecret(ctx, req.(*GetSecretRequest)) }
	return interceptor(ctx, in, info, h)
}

func _SecretsEnvService_ListSecrets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListSecretsRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(SecretsEnvServiceServer).ListSecrets(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/lma.secretsenv.v1.SecretsEnvService/ListSecrets"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(SecretsEnvServiceServer).ListSecrets(ctx, req.(*ListSecretsRequest)) }
	return interceptor(ctx, in, info, h)
}

func _SecretsEnvService_CheckEnvParity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckEnvParityRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(SecretsEnvServiceServer).CheckEnvParity(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity"}
	h := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(SecretsEnvServiceServer).CheckEnvParity(ctx, req.(*CheckEnvParityRequest)) }
	return interceptor(ctx, in, info, h)
}

var _SecretsEnvService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "lma.secretsenv.v1.SecretsEnvService",
	HandlerType: (*SecretsEnvServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetSecret", Handler: _SecretsEnvService_GetSecret_Handler},
		{MethodName: "ListSecrets", Handler: _SecretsEnvService_ListSecrets_Handler},
		{MethodName: "CheckEnvParity", Handler: _SecretsEnvService_CheckEnvParity_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "packages/contracts/proto/lma/secretsenv/v1/service.proto",
}
