package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/mcpserver"
)

func checkHealth(endpoint string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: HTTP %d", resp.StatusCode)
	}
	return nil
}

type bearerTransport struct{ token string }

func (t bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	clone := r.Clone(r.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(clone)
}

// smoke uses the same public MCP protocol as an agent, without requiring a specific desktop app.
func smoke(endpoint, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client := mcp.NewClient(&mcp.Implementation{Name: "webmaster-mcp-smoke", Version: mcpserver.ServerVersion}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endpoint, HTTPClient: &http.Client{Transport: bearerTransport{token}, Timeout: 45 * time.Second}, DisableStandaloneSSE: true}, nil)
	if err != nil {
		return err
	}
	defer func() { _ = session.Close() }()
	catalog, err := session.ListTools(ctx, nil)
	if err != nil {
		return err
	}
	if len(catalog.Tools) != mcpserver.ToolCount && len(catalog.Tools) != mcpserver.ToolCount-1 {
		return fmt.Errorf("expected %d tools, got %d", mcpserver.ToolCount, len(catalog.Tools))
	}
	for _, tool := range catalog.Tools {
		if tool.OutputSchema == nil || tool.Annotations == nil {
			return fmt.Errorf("missing output/annotation contract: %s", tool.Name)
		}
		if tool.Name == "submit_recrawl" {
			if tool.Annotations.ReadOnlyHint {
				return fmt.Errorf("submit_recrawl must not be read-only")
			}
			continue
		}
		if !tool.Annotations.ReadOnlyHint {
			return fmt.Errorf("missing read-only contract: %s", tool.Name)
		}
	}
	call := func(name string, args any) error {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			return err
		}
		if result.IsError {
			return fmt.Errorf("%s failed: %v", name, result.Content)
		}
		raw, err := json.Marshal(result.StructuredContent)
		if err != nil {
			return err
		}
		if string(raw) == "null" {
			return fmt.Errorf("%s has no structured result", name)
		}
		fmt.Println(name + ": ok")
		return nil
	}
	if err := call("get_user", map[string]any{}); err != nil {
		return err
	}
	if err := call("get_region_ids", map[string]any{}); err != nil {
		return err
	}
	if err := call("get_feed_regions", map[string]any{}); err != nil {
		return err
	}
	var hosts mcpserver.HostsOutput
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_hosts", Arguments: map[string]any{}})
	if err != nil {
		return err
	}
	if result.IsError {
		return fmt.Errorf("list_hosts failed: %v", result.Content)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &hosts); err != nil {
		return err
	}
	fmt.Println("list_hosts: ok")
	if len(hosts.Hosts) == 0 {
		return fmt.Errorf("no hosts: confirm the OAuth token can access Webmaster")
	}
	hostID := hosts.Hosts[0].HostID
	for _, name := range []string{"get_host", "get_summary", "get_diagnostics", "get_sitemaps", "get_recrawl_quota"} {
		if err := call(name, map[string]any{"host_id": hostID}); err != nil {
			return err
		}
	}
	return nil
}
