package goswag_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/diegoclair/goswag"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type greetIn struct {
	Name string `json:"name" jsonschema:"the name to greet"`
}

type greetOut struct {
	Message string `json:"message"`
}

// connect wires an in-memory client to the given server and returns the client session.
func connect(t *testing.T, ctx context.Context, srv *mcp.Server) *mcp.ClientSession {
	t.Helper()
	clientT, serverT := mcp.NewInMemoryTransports()
	_, err := srv.Connect(ctx, serverT, nil)
	require.NoError(t, err)
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func TestMCPServer_ToolRoundtrip(t *testing.T) {
	ctx := context.Background()
	var preCalled, postCalled bool

	srv := goswag.NewMCPServer("test-api", "0.0.1",
		[]goswag.MCPTool{
			goswag.Tool("greet", "Greet someone by name.", func(_ context.Context, in greetIn) (greetOut, error) {
				return greetOut{Message: "hello, " + in.Name}, nil
			}),
		},
		goswag.WithPreInvoke(func(_ context.Context, tool string, _ any) error {
			preCalled = true
			assert.Equal(t, "greet", tool)
			return nil
		}),
		goswag.WithPostInvoke(func(_ context.Context, _ string, _, _ any, _ error) { postCalled = true }),
	)

	cs := connect(t, ctx, srv)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "world"},
	})
	require.NoError(t, err)
	assert.False(t, res.IsError)

	// structured output is auto-serialized into the result; assert it flowed through.
	raw, _ := json.Marshal(res)
	assert.Contains(t, string(raw), "hello, world")
	assert.True(t, preCalled, "PreInvoke must run before the handler")
	assert.True(t, postCalled, "PostInvoke must run after the handler")
}

func TestMCPServer_PreInvokeRejects(t *testing.T) {
	ctx := context.Background()

	srv := goswag.NewMCPServer("test-api", "0.0.1",
		[]goswag.MCPTool{
			goswag.Tool("greet", "Greet someone by name.", func(_ context.Context, in greetIn) (greetOut, error) {
				return greetOut{Message: "hello, " + in.Name}, nil
			}),
		},
		goswag.WithPreInvoke(func(_ context.Context, _ string, _ any) error {
			return errors.New("denied by policy")
		}),
	)

	cs := connect(t, ctx, srv)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "greet", Arguments: map[string]any{"name": "world"}})
	// A handler error surfaces as a tool error result (not a transport error).
	if err == nil {
		assert.True(t, res.IsError, "a rejected PreInvoke must surface as a tool error")
		raw, _ := json.Marshal(res)
		assert.Contains(t, strings.ToLower(string(raw)), "denied by policy")
	}
}
