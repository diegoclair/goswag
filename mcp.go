package goswag

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPConfig holds the generic, content-agnostic hooks goswag wraps around every
// tool invocation. goswag takes no position on what they do — auth, rate
// limiting, tenant scoping and audit logging are the consumer's concern, wired
// through these. PreInvoke returning a non-nil error rejects the call before the
// handler runs.
type MCPConfig struct {
	PreInvoke  func(ctx context.Context, tool string, input any) error
	PostInvoke func(ctx context.Context, tool string, input, output any, err error)
}

// MCPOption configures an MCPConfig.
type MCPOption func(*MCPConfig)

// WithPreInvoke sets the pre-invocation gate (auth/rate-limit/etc.).
func WithPreInvoke(f func(ctx context.Context, tool string, input any) error) MCPOption {
	return func(c *MCPConfig) { c.PreInvoke = f }
}

// WithPostInvoke sets the post-invocation observer (audit/metrics/etc.).
func WithPostInvoke(f func(ctx context.Context, tool string, input, output any, err error)) MCPOption {
	return func(c *MCPConfig) { c.PostInvoke = f }
}

// MCPTool is an opaque, typed tool registration produced by Tool[In, Out].
// Go methods cannot have type parameters, so the In/Out types are captured in a
// closure here rather than on a chainable .AsTool[In,Out](...) method.
type MCPTool interface {
	register(s *mcp.Server, cfg *MCPConfig)
}

// Tool builds an MCP tool from a pure, typed handler. The input JSON Schema is
// inferred from In (with `jsonschema:` struct tags) and the output schema from
// Out — and the SDK validates incoming arguments against the input schema before
// the handler runs, so consumers do not hand-roll decoding/validation.
func Tool[In, Out any](name, description string, handler func(ctx context.Context, in In) (Out, error)) MCPTool {
	return &typedTool[In, Out]{name: name, description: description, handler: handler}
}

type typedTool[In, Out any] struct {
	name        string
	description string
	handler     func(ctx context.Context, in In) (Out, error)
}

func (t *typedTool[In, Out]) register(s *mcp.Server, cfg *MCPConfig) {
	mcp.AddTool(s, &mcp.Tool{Name: t.name, Description: t.description},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
			if cfg.PreInvoke != nil {
				if err := cfg.PreInvoke(ctx, t.name, in); err != nil {
					var zero Out
					return nil, zero, err
				}
			}
			out, err := t.handler(ctx, in)
			if cfg.PostInvoke != nil {
				cfg.PostInvoke(ctx, t.name, in, out, err)
			}
			return nil, out, err
		})
}

// NewMCPServer builds an *mcp.Server from typed tools plus the generic hooks.
// The caller serves it over the SDK's stdio or Streamable HTTP transports (the
// latter mounts into an existing HTTP mux — see the design doc §5).
func NewMCPServer(name, version string, tools []MCPTool, opts ...MCPOption) *mcp.Server {
	cfg := &MCPConfig{}
	for _, o := range opts {
		o(cfg)
	}
	s := mcp.NewServer(&mcp.Implementation{Name: name, Version: version}, nil)
	for _, t := range tools {
		t.register(s, cfg)
	}
	return s
}
