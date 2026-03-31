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
	case "unique_values":
		return "Collect unique values from a fixed field in each line (naive delimiter split; not CSV-safe).", map[string]any{
			"type":     "object",
			"required": []string{"path", "field_index"},
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Relative path under root.",
				},
				"root": map[string]any{
					"type":        "string",
					"description": "Sandbox root; defaults to cwd.",
				},
				"delimiter": map[string]any{
					"type":        "string",
					"description": "Field separator; default comma; max 3 bytes. No quoting rules.",
				},
				"field_index": map[string]any{
					"type":        "integer",
					"description": "0-based column index; max 1000.",
				},
				"trim_space": map[string]any{
					"type":        "boolean",
					"description": "Trim ASCII space around field; default true.",
				},
				"max_unique": map[string]any{
					"type":        "integer",
					"description": "Stop after this many distinct values (scan order); default 1000; hard max 10000.",
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": "Max lines to read; default 1_000_000; hard max 1_000_000.",
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
