package tools

import (
	"context"
	"encoding/json"
)

// PingHandler implements server.ping → {"ok": true}.
func PingHandler(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]bool{"ok": true})
}
