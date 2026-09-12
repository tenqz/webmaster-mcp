package mcpserver_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/config"
	"github.com/tenqz/yandex-webmaster-mcp/internal/httpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/mcpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
)

type auditRoundTrip func(*http.Request) (*http.Response, error)

func (f auditRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestEndToEnd verifies HTTP/auth/MCP/Webmaster against an in-memory Yandex transport.
func TestEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var apiCalls atomic.Int32
	upstream := &http.Client{Transport: auditRoundTrip(func(r *http.Request) (*http.Response, error) {
		apiCalls.Add(1)
		if got := r.Header.Get("Authorization"); got != "OAuth yandex-token" {
			t.Errorf("Authorization=%q", got)
		}
		body := `{"user_id":1,"hosts":[{"host_id":"https:example.com:443","verified":true}]}`
		if strings.Contains(r.URL.Path, "/summary") {
			body = `{"sqi":10}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})}
	api := webmaster.NewClient(upstream, "yandex-token", "https://api.webmaster.yandex.net/v4", webmaster.Options{})
	server := mcpserver.New(api)
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true})
	ts := httptest.NewServer(httpserver.New(config.Config{MCPPath: "/mcp", AuthToken: "secret"}, handler))
	defer ts.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "e2e", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL + "/mcp", HTTPClient: &http.Client{Transport: roundTripHeader("secret")}, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_hosts", Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatalf("list_hosts: %v %#v", err, result)
	}
	raw, _ := json.Marshal(result.StructuredContent)
	if !strings.Contains(string(raw), "https:example.com:443") {
		t.Fatalf("payload=%s", raw)
	}
	if apiCalls.Load() == 0 {
		t.Fatal("expected Yandex calls")
	}
}

type headerTransport struct{ token string }

func roundTripHeader(token string) http.RoundTripper { return headerTransport{token} }

func (t headerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	clone := r.Clone(r.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(clone)
}
