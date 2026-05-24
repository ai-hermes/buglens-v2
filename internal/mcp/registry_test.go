package mcp

import (
	"testing"

	"github.com/ai-hermes/buglens-v2/internal/config"
)

func TestToolRegistryCountAndNames(t *testing.T) {
	r := NewRegistry(config.Config{})
	names := r.ToolNames()
	if len(names) != 36 {
		t.Fatalf("expected 36 tools, got %d", len(names))
	}
	mustHave := []string{"gitlab_list_projects", "arms_rum_list_apps", "arms_rum_search_errors"}
	for _, n := range mustHave {
		found := false
		for _, got := range names {
			if got == n {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing tool: %s", n)
		}
	}
	for _, n := range names {
		if len(n) >= 8 && n[:8] == "buglens_" {
			t.Fatalf("tool should be unprefixed: %s", n)
		}
	}
}
