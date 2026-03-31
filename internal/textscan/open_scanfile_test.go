package textscan

import (
	"os"
	"path/filepath"
	"testing"

	"wc-stats-mcp/internal/validate"
)

func TestScanLinesFile_integration(t *testing.T) {
	root := t.TempDir()
	rel := "x.txt"
	if err := os.WriteFile(filepath.Join(root, rel), []byte("alpha\nbeta\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var lines int
	err := ScanLinesFile(root, rel, validate.Limits{MaxFileBytes: 1 << 20}, ScanLimits{MaxLines: 100, MaxLineBytes: 1024},
		func(line []byte, lineNo int) error {
			lines++
			if lineNo != lines {
				t.Fatalf("lineNo %d lines %d", lineNo, lines)
			}
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if lines != 2 {
		t.Fatalf("lines %d", lines)
	}
}
