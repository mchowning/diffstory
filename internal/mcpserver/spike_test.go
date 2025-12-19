//go:build spike

package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestSDKSpike verifies the MCP SDK API matches our assumptions.
// Run with: nix develop -c go test -tags=spike ./internal/mcpserver/...
func TestSDKSpike(t *testing.T) {
	// Verify server creation
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	// Verify tool registration compiles
	// This tests that our expected handler signature is correct
	// Note: jsonschema tag format is just "description" not "key=value"
	type TestInput struct {
		Message string `json:"message" jsonschema:"A test message"`
	}

	type TestOutput struct {
		Response string `json:"response" jsonschema:"The response message"`
	}

	// Handler with structured output (typed second return)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TestInput) (*mcp.CallToolResult, *TestOutput, error) {
		return nil, &TestOutput{Response: "Hello " + input.Message}, nil
	})

	// Handler with content-based output (using CallToolResult)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "content_tool",
		Description: "A tool that returns content",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TestInput) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Hello " + input.Message},
			},
		}, nil, nil
	})

	// Verify StdioTransport exists
	// Note: We can't actually run the server in a test without blocking
	_ = &mcp.StdioTransport{}

	t.Log("SDK API verified successfully")
}
