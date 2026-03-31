package tools

// ToolMeta returns MCP list metadata for a registered tool name.
func ToolMeta(name string) (description string, schema map[string]any) {
	switch name {
	case "server.ping":
		return "Liveness check.", map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	case "count_lines_words_bytes":
		return "Count lines, words, and bytes in a file using coreutils wc.", map[string]any{
			"type":     "object",
			"required": []string{"path"},
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Relative path under root.",
				},
				"root": map[string]any{
					"type":        "string",
					"description": "Sandbox root directory; defaults to the server process current working directory.",
				},
			},
		}
	default:
		return "Tool.", map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}
}
