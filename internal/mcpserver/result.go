package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
)

func jsonResult(value any) (*mcp.CallToolResult, any, error) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("encode tool result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(body)}},
	}, value, nil
}

func toolError(format string, args ...any) (*mcp.CallToolResult, any, error) {
	return toolFailure(fmt.Errorf(format, args...))
}

func toolFailure(err error) (*mcp.CallToolResult, any, error) {
	payload := map[string]any{"code": "invalid_argument", "message": err.Error(), "retryable": false}
	var failure *webmaster.RequestError
	if errors.As(err, &failure) {
		payload["code"], payload["retryable"] = failure.Kind, failure.Retryable
		if failure.Status != 0 {
			payload["httpStatus"] = failure.Status
		}
		if failure.RetryAfterSeconds > 0 {
			payload["retryAfterSeconds"] = failure.RetryAfterSeconds
		}
	}
	wrapped := map[string]any{"error": payload}
	raw, _ := json.Marshal(wrapped)
	return &mcp.CallToolResult{IsError: true, StructuredContent: wrapped, Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}, nil, nil
}

func observe[In, Out any](name string, next mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		start := time.Now()
		result, out, err := next(ctx, req, in)
		failed := err != nil || (result != nil && result.IsError)
		slog.Info("mcp_tool", "tool", name, "failed", failed, "duration_ms", time.Since(start).Milliseconds())
		return result, out, err
	}
}

func schemaFor[T any]() *jsonschema.Schema {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		return &jsonschema.Schema{Type: "object"}
	}
	return schema
}

func objectSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "object", AdditionalProperties: &jsonschema.Schema{}}
}

// reportSchema documents stable fields while preserving upstream extensions and missing values.
func reportSchema(name string) *jsonschema.Schema {
	s := objectSchema()
	number := func(description string) *jsonschema.Schema {
		return &jsonschema.Schema{Type: "number", Description: description}
	}
	switch name {
	case "get_summary":
		s.Properties = map[string]*jsonschema.Schema{
			"sqi":                    number("Site Quality Index; absent means unavailable"),
			"searchable_pages_count": number("Pages in search; count, not traffic"),
			"excluded_pages_count":   number("Excluded pages count"),
			"site_problems":          objectSchema(),
		}
	case "get_query_history":
		s.Properties = map[string]*jsonschema.Schema{"indicators": {Type: "object", Description: "Source indicator names mapped to dated observations; preserve source dates and missing indicators", AdditionalProperties: &jsonschema.Schema{Type: "array", Items: objectSchema()}}}
	case "get_popular_queries":
		s.Properties = map[string]*jsonschema.Schema{"count": number("Provider count, not a guarantee of exhaustive query coverage"), "queries": {Type: "array", Items: objectSchema()}}
	case "get_diagnostics":
		s.Properties = map[string]*jsonschema.Schema{"problems": objectSchema()}
	}
	return s
}
