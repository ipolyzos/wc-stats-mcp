package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRegistry_routesByName(t *testing.T) {
	r := NewRegistry()
	r.Register("server.ping", PingHandler)

	got, err := r.Call(context.Background(), "server.ping", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]bool
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatal(err)
	}
	if !payload["ok"] {
		t.Fatalf("expected ok true, got %v", payload)
	}
}

func TestRegistry_unknownTool(t *testing.T) {
	r := NewRegistry()
	_, err := r.Call(context.Background(), "nope", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
