package tools

import (
	"fmt"
	"strings"
)

// ParseWCCounts parses stdout from `wc -l -w -c` (one line per file, optional filename suffix).
// It reads the first three integers from the first non-empty line, ignoring extra whitespace and trailing path.
func ParseWCCounts(stdout []byte) (lines, words, byteCount int, err error) {
	s := string(stdout)
	for _, raw := range strings.Split(s, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		return parseWCCountsLine(line)
	}
	return 0, 0, 0, fmt.Errorf("wc: empty stdout")
}

func parseWCCountsLine(line string) (lines, words, byteCount int, err error) {
	// Leading spaces, variable-width columns, optional filename: "  10  20  120  /path/to/file"
	var l, w, b int
	n, scanErr := fmt.Sscanf(strings.TrimSpace(line), "%d %d %d", &l, &w, &b)
	if n != 3 || scanErr != nil {
		return 0, 0, 0, fmt.Errorf("wc: cannot parse counts from %q: %w", line, scanErr)
	}
	return l, w, b, nil
}
