package mcpserver_test

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/mcpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
)

func text(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected text content")
	}
	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	return content.Text
}

// TestGetUserReturnsJSON documents the get_user payload shape for agents.
func TestGetUserReturnsJSON(t *testing.T) {
	tools := &mcpserver.Toolset{API: webmaster.NewDemo()}
	result, _, err := tools.GetUser(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text(t, result), `"user_id"`) {
		t.Fatalf("payload = %s", text(t, result))
	}
}

// TestGetSummaryRejectsMissingHost documents tool-level validation errors are MCP errors.
func TestGetSummaryRejectsMissingHost(t *testing.T) {
	tools := &mcpserver.Toolset{API: webmaster.NewDemo()}
	result, _, err := tools.GetSummary(context.Background(), nil, mcpserver.HostInput{})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected a tool error")
	}
}
