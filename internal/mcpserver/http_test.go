package mcpserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/mcpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
)

// TestStreamableHTTPExposesCatalog documents that a remote agent can list Webmaster tools over HTTP.
func TestStreamableHTTPExposesCatalog(t *testing.T) {
	server := mcpserver.New(webmaster.NewDemo())
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(listed.Tools) != mcpserver.ToolCount {
		t.Fatalf("got %d tools, want %d", len(listed.Tools), mcpserver.ToolCount)
	}
	found := map[string]bool{}
	for _, tool := range listed.Tools {
		found[tool.Name] = true
		if tool.OutputSchema == nil || tool.Annotations == nil {
			t.Fatalf("missing contract: %s", tool.Name)
		}
		if tool.Name == "submit_recrawl" {
			if tool.Annotations.ReadOnlyHint {
				t.Fatal("submit_recrawl must not be read-only")
			}
			continue
		}
		if !tool.Annotations.ReadOnlyHint {
			t.Fatalf("%s should be read-only", tool.Name)
		}
	}
	for _, name := range []string{"list_hosts", "get_popular_queries", "submit_recrawl", "get_region_ids"} {
		if !found[name] {
			t.Fatalf("missing tool %s", name)
		}
	}
}
