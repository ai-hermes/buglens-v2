package monitoring

import (
	"strings"
	"testing"
)

func TestParseLastDurationMS(t *testing.T) {
	v, _ := ParseLastDurationMS("15m")
	if v != 15*60_000 {
		t.Fatalf("got %d", v)
	}
}

func TestBuildRUMSearchQueryStructured(t *testing.T) {
	q := BuildRUMSearchQuery("", "exception", "a@b", []string{"browser", "miniapp"}, `RUM_UNHANDLED_REJECTION: "boom"`, "TypeError")
	if q == "" {
		t.Fatalf("empty query")
	}
	if !strings.Contains(q, `app.id : "a@b"`) || !strings.Contains(q, "event_type: exception") {
		t.Fatalf("query mismatch: %s", q)
	}
}
