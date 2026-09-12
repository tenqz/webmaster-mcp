package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/config"
	"github.com/tenqz/yandex-webmaster-mcp/internal/httpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/mcpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
)

// TestDemoSmoke exercises the release check through the real HTTP handler with public fixtures.
func TestDemoSmoke(t *testing.T) {
	server := mcpserver.New(webmaster.NewDemo())
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true})
	ts := httptest.NewServer(httpserver.New(config.Config{MCPPath: "/mcp", AuthToken: "test"}, handler))
	defer ts.Close()
	if err := checkHealth(ts.URL + "/health"); err != nil {
		t.Fatal(err)
	}
	if err := smoke(ts.URL+"/mcp", "test"); err != nil {
		t.Fatal(err)
	}
	if err := smoke(ts.URL+"/mcp", "wrong"); err == nil {
		t.Fatal("wrong token accepted")
	}
}
