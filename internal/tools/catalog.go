package tools

// Shared inputSchema copy for path-based tools (tool-contract.md).
const (
	schemaDescPath     = "Relative path under root; must be relative with no .. path elements."
	schemaDescRoot     = "Sandbox root directory used to resolve path; defaults to the server process current working directory."
	schemaDescMaxBytes = "Optional stricter maximum resolved file size in bytes; cannot exceed the server effective cap. Omit or use 0 for the server default (see tool-contract)."
	schemaDescMaxLines = "Maximum lines to read from the file; default 1_000_000; hard maximum 1_000_000."
)

func propPath() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": schemaDescPath,
	}
}

func propRoot() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": schemaDescRoot,
	}
}

func propMaxBytes() map[string]any {
	return map[string]any{
		"type":        "integer",
		"description": schemaDescMaxBytes,
	}
}

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
				"path":      propPath(),
				"root":      propRoot(),
				"max_bytes": propMaxBytes(),
			},
		}
	case "top_terms":
		return "Stream token frequencies from a text file (Unicode letter/digit tokens) with line and size caps.", map[string]any{
			"type":     "object",
			"required": []string{"path"},
			"properties": map[string]any{
				"path":      propPath(),
				"root":      propRoot(),
				"max_bytes": propMaxBytes(),
				"top_n": map[string]any{
					"type":        "integer",
					"description": "Maximum number of terms to return; default 20; maximum 100.",
				},
				"offset": map[string]any{
					"type":        "integer",
					"description": "Start index into the sorted terms list for pagination; default 0.",
				},
				"min_len": map[string]any{
					"type":        "integer",
					"description": "Minimum token length in Unicode runes; default 3; maximum 30.",
				},
				"case_sensitive": map[string]any{
					"type":        "boolean",
					"description": "If false (default), normalize the line with bytes.ToLower before tokenizing.",
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": schemaDescMaxLines,
				},
			},
		}
	case "unique_values":
		return "Stream delimited lines and return sorted distinct values for one field index (naive split; not CSV-safe).", map[string]any{
			"type":     "object",
			"required": []string{"path", "field_index"},
			"properties": map[string]any{
				"path":      propPath(),
				"root":      propRoot(),
				"max_bytes": propMaxBytes(),
				"delimiter": map[string]any{
					"type":        "string",
					"description": "Field delimiter; default comma; maximum 3 bytes.",
				},
				"field_index": map[string]any{
					"type":        "integer",
					"description": "Zero-based field index; maximum 1000.",
				},
				"trim_space": map[string]any{
					"type":        "boolean",
					"description": "Trim ASCII spaces around the field; default true when omitted.",
				},
				"max_unique": map[string]any{
					"type":        "integer",
					"description": "Stop after this many distinct values; default 1000; hard maximum 10_000.",
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": schemaDescMaxLines,
				},
			},
		}
	case "field_frequency":
		return "Stream delimited lines and return top-N value frequencies for one field (naive split; not CSV-safe).", map[string]any{
			"type":     "object",
			"required": []string{"path", "field_index"},
			"properties": map[string]any{
				"path":      propPath(),
				"root":      propRoot(),
				"max_bytes": propMaxBytes(),
				"delimiter": map[string]any{
					"type":        "string",
					"description": "Field delimiter; default comma; maximum 3 bytes.",
				},
				"field_index": map[string]any{
					"type":        "integer",
					"description": "Zero-based field index; maximum 1000.",
				},
				"trim_space": map[string]any{
					"type":        "boolean",
					"description": "Trim ASCII spaces around the field; default true when omitted.",
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": schemaDescMaxLines,
				},
				"top_n": map[string]any{
					"type":        "integer",
					"description": "Maximum frequency rows to return; default 20; maximum 100.",
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
