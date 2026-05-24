package arms

import (
	"context"
	"testing"

	"github.com/ai-hermes/buglens-v2/internal/config"
	gmcp "github.com/mark3labs/mcp-go/mcp"
)

func TestArmsRumSearchErrorsUsesARMSPath(t *testing.T) {
	t.Setenv("BUGLENS_MONITORING_MOCK", "true")

	handlers := map[string]func(context.Context, map[string]any) (map[string]any, error){}
	RegisterMCPTools(config.Config{}, func(name, _ string, handler func(context.Context, map[string]any) (map[string]any, error), _ ...gmcp.ToolOption) {
		handlers[name] = handler
	})

	h := handlers["arms_rum_search_errors"]
	if h == nil {
		t.Fatalf("handler not registered")
	}

	payload, err := h(context.Background(), map[string]any{
		"last":      "15m",
		"page_size": 20,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if payload["success"] != true {
		t.Fatalf("expected success payload")
	}
	query, ok := payload["query"].(string)
	if !ok || query == "" {
		t.Fatalf("missing query in payload")
	}
	data := payload["data"].(map[string]any)
	if _, ok := data["items"].([]map[string]any); !ok {
		t.Fatalf("items not present or malformed: %#v", data["items"])
	}
}

func TestArmsRumSearchErrorsExplicitTimeAndQuery(t *testing.T) {
	t.Setenv("BUGLENS_MONITORING_MOCK", "true")

	handlers := map[string]func(context.Context, map[string]any) (map[string]any, error){}
	RegisterMCPTools(config.Config{}, func(name, _ string, handler func(context.Context, map[string]any) (map[string]any, error), _ ...gmcp.ToolOption) {
		handlers[name] = handler
	})

	h := handlers["arms_rum_search_errors"]
	payload, err := h(context.Background(), map[string]any{
		"time_from_ms": 1779428866000,
		"time_to_ms":   1779429946000,
		"query":        "custom-query",
		"page_token":   "2",
		"page_size":    100,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if payload["query"] != "custom-query" {
		t.Fatalf("query not echoed: %#v", payload["query"])
	}
	if payload["success"] != true {
		t.Fatalf("expected success")
	}
	if _, ok := payload["next_page_token"].(string); !ok {
		t.Fatalf("next_page_token should be present")
	}
}
