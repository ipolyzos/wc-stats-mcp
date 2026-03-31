package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wc-stats-mcp/internal/validate"
)

func TestUniqueValues_golden(t *testing.T) {
	root := t.TempDir()
	raw, err := os.ReadFile("testdata/unique_values_sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data.csv"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	h := UniqueValuesHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, err := json.Marshal(map[string]any{
		"path":        "data.csv",
		"root":        root,
		"field_index": 1,
		"delimiter":   ",",
		"trim_space":  true,
		"max_unique":  5000,
		"max_lines":   1_000_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		FieldIndex int      `json:"field_index"`
		Values     []string `json:"values"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got.FieldIndex != 1 {
		t.Fatalf("field_index %d", got.FieldIndex)
	}
	want := []string{"admin", "guest", "user"}
	if len(got.Values) != len(want) {
		t.Fatalf("got %v", got.Values)
	}
	for i := range want {
		if got.Values[i] != want[i] {
			t.Fatalf("idx %d: got %q want %q", i, got.Values[i], want[i])
		}
	}
}

func TestUniqueValues_shortRowsSkipped(t *testing.T) {
	root := t.TempDir()
	// Only one column — field_index 2 missing
	content := "only\na,b\n"
	if err := os.WriteFile(filepath.Join(root, "x.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	h := UniqueValuesHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":        "x.txt",
		"root":        root,
		"field_index": 2,
		"delimiter":   ",",
	})
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Values []string `json:"values"`
	}
	_ = json.Unmarshal(out, &got)
	// "a,b" gives field 2 as missing? parts ["a","b"] idx 2 -> skip; "only" -> skip
	if len(got.Values) != 0 {
		t.Fatalf("expected no values, got %v", got.Values)
	}
}

func TestUniqueValues_pipeDelimiter(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "p.txt"), []byte("x|y|z\na|b|c\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	h := UniqueValuesHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":        "p.txt",
		"root":        root,
		"field_index": 2,
		"delimiter":   "|",
	})
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Values []string `json:"values"`
	}
	_ = json.Unmarshal(out, &got)
	if len(got.Values) != 2 || got.Values[0] != "c" || got.Values[1] != "z" {
		t.Fatalf("%v", got.Values)
	}
}

func TestUniqueValues_maxUniqueStopsEarly(t *testing.T) {
	root := t.TempDir()
	var b []byte
	for i := range 50 {
		b = append(b, []byte(string(rune('a'+i)))...)
		b = append(b, '\n')
	}
	if err := os.WriteFile(filepath.Join(root, "many.txt"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	h := UniqueValuesHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":        "many.txt",
		"root":        root,
		"field_index": 0,
		"delimiter":   ",",
		"max_unique":  3,
	})
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Values []string `json:"values"`
	}
	_ = json.Unmarshal(out, &got)
	if len(got.Values) != 3 {
		t.Fatalf("got %d values %v", len(got.Values), got.Values)
	}
}

func TestUniqueValues_rejectsLargeDelimiter(t *testing.T) {
	h := UniqueValuesHandler(validate.Limits{})
	args, _ := json.Marshal(map[string]any{
		"path":        "x",
		"root":        t.TempDir(),
		"field_index": 0,
		"delimiter":   "abcd",
	})
	_, err := h(context.Background(), args)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUniqueValues_requiresFieldIndex(t *testing.T) {
	h := UniqueValuesHandler(validate.Limits{})
	args, _ := json.Marshal(map[string]any{"path": "x", "root": t.TempDir()})
	_, err := h(context.Background(), args)
	if err == nil || !strings.Contains(err.Error(), "field_index") {
		t.Fatalf("got %v", err)
	}
}
