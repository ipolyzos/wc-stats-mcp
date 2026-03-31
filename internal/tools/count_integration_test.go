package tools

import (
	"context"
	"encoding/json"
	"os"
	osexec "os/exec"
	"path/filepath"
	"testing"

	"wc-stats-mcp/internal/exec"
	"wc-stats-mcp/internal/validate"
)

func TestIntegration_countLinesWordsBytes_wc(t *testing.T) {
	if _, err := osexec.LookPath("wc"); err != nil {
		t.Skip("wc not on PATH:", err)
	}

	root := t.TempDir()
	rel := "sample.txt"
	// Two lines, two words, 12 bytes.
	content := "hello\nworld\n"
	if err := os.WriteFile(filepath.Join(root, rel), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	h := CountLinesWordsBytesHandler(exec.NewRunner(), validate.Limits{MaxFileBytes: 1 << 20})
	args, err := json.Marshal(map[string]string{
		"path": rel,
		"root": root,
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Path  string `json:"path"`
		Lines int    `json:"lines"`
		Words int    `json:"words"`
		Bytes int    `json:"bytes"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Lines != 2 || got.Words != 2 || got.Bytes != 12 {
		t.Fatalf("unexpected counts: %+v", got)
	}
	if got.Path == "" {
		t.Fatal("empty path")
	}
}
