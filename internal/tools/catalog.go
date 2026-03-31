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
	case "unique_values":
		return "Stream delimited lines and return sorted distinct values for one field index (naive split; not CSV-safe).", map[string]any{
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
					"description": "Field delimiter; default comma; max 3 bytes.",
				},
				"field_index": map[string]any{
					"type":        "integer",
					"description": "0-based column index; max 1000.",
				},
				"trim_space": map[string]any{
					"type":        "boolean",
					"description": "Trim ASCII space around the field; default true.",
				},
				"max_unique": map[string]any{
					"type":        "integer",
					"description": "Stop after this many distinct values; default 1000; hard max 10_000.",
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": "Max lines to scan; default 1_000_000; hard max 1_000_000.",
				},
			},
		}
	case "field_frequency":
		return "Stream delimited lines and return top-N value frequencies for one field (naive split; not CSV-safe).", map[string]any{
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
					"description": "Field delimiter; default comma; max 3 bytes.",
				},
				"field_index": map[string]any{
					"type":        "integer",
					"description": "0-based column index; max 1000.",
				},
				"trim_space": map[string]any{
					"type":        "boolean",
					"description": "Trim ASCII space around the field; default true.",
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": "Max lines to scan; default 1_000_000; hard max 1_000_000.",
				},
				"top_n": map[string]any{
					"type":        "integer",
					"description": "Max frequency rows to return; default 20; max 100.",
				},
				"include_empty": map[string]any{
					"type":        "boolean",
					"description": "Count empty field values after trim; default false.",
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
