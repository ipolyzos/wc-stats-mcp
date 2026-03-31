package tools

import (
	"wc-stats-mcp/internal/exec"
	"wc-stats-mcp/internal/validate"
)

// ProductionToolNames is the sorted tool names RegisterProduction adds. Keep in sync with RegisterProduction.
var ProductionToolNames = []string{
	"count_lines_words_bytes",
	"server.ping",
	"top_terms",
	"unique_values",
}

// RegisterProduction registers the same handlers as cmd/wc-stats-mcp.
func RegisterProduction(reg *Registry, r *exec.Runner, lim validate.Limits) {
	if reg == nil {
		return
	}
	if r == nil {
		r = exec.NewRunner()
	}
	reg.Register("server.ping", PingHandler)
	reg.Register("count_lines_words_bytes", CountLinesWordsBytesHandler(r, lim))
	reg.Register("top_terms", TopTermsHandler(lim))
	reg.Register("unique_values", UniqueValuesHandler(lim))
}
