package mcpserver_test

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenqz/yandex-webmaster-mcp/internal/mcpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadOnlyServerRejectsDirectRecrawl(t *testing.T) {
	server := mcpserver.New(webmaster.NewDemo(), true)
	ts := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true}))
	defer ts.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()
	list, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != 33 {
		t.Fatalf("got %d tools", len(list.Tools))
	}
	for _, tool := range list.Tools {
		if tool.Name == "submit_recrawl" {
			t.Fatal("write exposed")
		}
	}
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "submit_recrawl", Arguments: map[string]any{"host_id": "https:example.invalid:443", "url": "https://example.invalid/"}})
	if err == nil && !result.IsError {
		t.Fatal("direct write succeeded")
	}
}
