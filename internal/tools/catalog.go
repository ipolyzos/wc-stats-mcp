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
	case "top_terms":
		return "Stream token frequencies from a text file (Unicode letter/digit tokens) with line and size caps.", map[string]any{
			"type":     "object",
			"required": []string{"path"},
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Relative path under root.",
				},
				"root": map[string]any{
					"type":        "string",
					"description": "Sandbox root; defaults to cwd.",
				},
				"top_n": map[string]any{
					"type":        "integer",
					"description": "Max terms to return; default 20, max 100.",
				},
				"min_len": map[string]any{
					"type":        "integer",
					"description": "Minimum token length in Unicode runes; default 3, max 30.",
				},
				"case_sensitive": map[string]any{
					"type":        "boolean",
					"description": "If false (default), fold case with Unicode lowercasing before counting.",
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": "Max lines to scan; default 1_000_000; hard-capped at 1_000_000.",
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
