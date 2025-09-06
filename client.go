package xai

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"time"

	v1 "github.com/bibyzan/xai-proto/gen/go/github.com/bibyzan/xai-proto/gen/go/xai/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	apiKey string
	auth   v1.AuthClient
	chat   v1.ChatClient
	conn   *grpc.ClientConn
}

// perRPCCreds injects API key and optional metadata on every RPC.
// It requires transport security unless explicitly overridden via WithInsecure.
type perRPCCreds struct {
	apiKey  string
	extraMD map[string]string
	secure  bool
}

func (c perRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	md := make(map[string]string, len(c.extraMD)+1)
	for k, v := range c.extraMD {
		md[k] = v
	}
	// Use a conventional header; adjust if the server expects a different key.
	md["x-api-key"] = c.apiKey
	return md, nil
}

func (c perRPCCreds) RequireTransportSecurity() bool { return c.secure }

// TimeoutInterceptor returns a unary interceptor that applies a per-call timeout.
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewClient creates a client using only an API key, with sensible defaults and environment overrides.
// Defaults:
//   - API host: "api.x.ai:443" (override with env XAI_API_HOST)
//   - Timeout: 60s (override with env XAI_TIMEOUT, e.g., "30s", "2m")
//   - TLS enabled (override with env XAI_INSECURE=true for local/dev)
func NewClient(apiKey string) *Client {
	host := os.Getenv("XAI_API_HOST")
	if host == "" {
		host = "api.x.ai:443"
	}

	to := 60 * time.Second
	if s := os.Getenv("XAI_TIMEOUT"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			to = d
		}
	}

	insecure := false
	if v := strings.ToLower(os.Getenv("XAI_INSECURE")); v == "1" || v == "true" || v == "yes" {
		insecure = true
	}

	return NewClientWithOptions(apiKey, host, nil, to, insecure)
}

// NewClientWithOptions creates a secure gRPC client using TLS, attaches API key and metadata on each call,
// and applies a timeout interceptor similar to the provided Python snippet.
//
// apiHost examples: "api.example.com:443" or "localhost:8080"
// metadata: optional extra headers to send with each request (e.g., {"x-client": "xai-go"}).
// timeout: per-RPC timeout applied via interceptor; use 0 for no timeout.
// useInsecure: set true only for local/dev without TLS.
func NewClientWithOptions(apiKey, apiHost string, metadata map[string]string, timeout time.Duration, useInsecure bool) *Client {
	// Transport credentials (TLS or Insecure for local testing only)
	var dialCreds grpc.DialOption
	prc := perRPCCreds{apiKey: apiKey, extraMD: metadata, secure: !useInsecure}
	if useInsecure {
		dialCreds = grpc.WithTransportCredentials(insecure.NewCredentials())
		// When insecure, also mark per-RPC creds as not requiring TLS
		prc.secure = false
	} else {
		dialCreds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}

	// Interceptors: timeout similar to grpc.intercept_channel(..., TimeoutInterceptor(timeout))
	opts := []grpc.DialOption{
		dialCreds,
		grpc.WithPerRPCCredentials(prc),
		grpc.WithChainUnaryInterceptor(TimeoutInterceptor(timeout)),
	}

	grpcConn, err := grpc.Dial(apiHost, opts...)
	if err != nil {
		panic(err)
	}

	return &Client{
		apiKey: apiKey,
		auth:   v1.NewAuthClient(grpcConn),
		chat:   v1.NewChatClient(grpcConn),
		conn:   grpcConn,
	}
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
