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

func TestFieldFrequency_goldenCountDescThenValueAsc(t *testing.T) {
	root := t.TempDir()
	raw, err := os.ReadFile("testdata/field_frequency_order.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "data.txt"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	h := FieldFrequencyHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, err := json.Marshal(map[string]any{
		"path":        "data.txt",
		"root":        root,
		"field_index": 1,
		"delimiter":   ",",
		"top_n":       100,
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
		Frequencies []struct {
			Value string `json:"value"`
			Count int    `json:"count"`
		} `json:"frequencies"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	want := []struct {
		Value string
		Count int
	}{
		{"a", 3},
		{"b", 2},
		{"c", 1},
	}
	if len(got.Frequencies) != len(want) {
		t.Fatalf("got %+v", got.Frequencies)
	}
	for i := range want {
		if got.Frequencies[i].Value != want[i].Value || got.Frequencies[i].Count != want[i].Count {
			t.Fatalf("idx %d: got %q:%d want %q:%d", i,
				got.Frequencies[i].Value, got.Frequencies[i].Count, want[i].Value, want[i].Count)
		}
	}
}

func TestFieldFrequency_goldenTieBreakValueAsc(t *testing.T) {
	root := t.TempDir()
	raw, err := os.ReadFile("testdata/field_frequency_tie.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "t.txt"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	h := FieldFrequencyHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":        "t.txt",
		"root":        root,
		"field_index": 0,
		"delimiter":   ",",
		"top_n":       10,
	})
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Frequencies []struct {
			Value string `json:"value"`
			Count int    `json:"count"`
		} `json:"frequencies"`
	}
	_ = json.Unmarshal(out, &got)
	if len(got.Frequencies) != 3 {
		t.Fatalf("got %+v", got.Frequencies)
	}
	// All count 1 → UTF-8 byte order: "a", "m", "z"
	if got.Frequencies[0].Value != "a" || got.Frequencies[1].Value != "m" || got.Frequencies[2].Value != "z" {
		t.Fatalf("order %+v", got.Frequencies)
	}
}

func TestFieldFrequency_topNTruncates(t *testing.T) {
	root := t.TempDir()
	raw, _ := os.ReadFile("testdata/field_frequency_order.txt")
	_ = os.WriteFile(filepath.Join(root, "data.txt"), raw, 0o600)
	h := FieldFrequencyHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":        "data.txt",
		"root":        root,
		"field_index": 1,
		"delimiter":   ",",
		"top_n":       2,
	})
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		TopN        int `json:"top_n"`
		Frequencies []struct {
			Value string `json:"value"`
			Count int    `json:"count"`
		} `json:"frequencies"`
	}
	_ = json.Unmarshal(out, &got)
	if got.TopN != 2 || len(got.Frequencies) != 2 {
		t.Fatalf("top_n=%d len=%d %+v", got.TopN, len(got.Frequencies), got.Frequencies)
	}
	if got.Frequencies[0].Value != "a" || got.Frequencies[1].Value != "b" {
		t.Fatalf("%+v", got.Frequencies)
	}
}

func TestFieldFrequency_excludeEmptyDefault(t *testing.T) {
	root := t.TempDir()
	raw, _ := os.ReadFile("testdata/field_frequency_empty.txt")
	_ = os.WriteFile(filepath.Join(root, "e.txt"), raw, 0o600)
	h := FieldFrequencyHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":        "e.txt",
		"root":        root,
		"field_index": 1,
		"delimiter":   ",",
		"top_n":       20,
	})
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Frequencies []struct {
			Value string `json:"value"`
			Count int    `json:"count"`
		} `json:"frequencies"`
	}
	_ = json.Unmarshal(out, &got)
	if len(got.Frequencies) != 1 || got.Frequencies[0].Value != "1" || got.Frequencies[0].Count != 1 {
		t.Fatalf("%+v", got.Frequencies)
	}
}

func TestFieldFrequency_includeEmpty(t *testing.T) {
	root := t.TempDir()
	raw, _ := os.ReadFile("testdata/field_frequency_empty.txt")
	_ = os.WriteFile(filepath.Join(root, "e.txt"), raw, 0o600)
	h := FieldFrequencyHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":          "e.txt",
		"root":          root,
		"field_index":   1,
		"delimiter":     ",",
		"top_n":         20,
		"include_empty": true,
	})
	out, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Frequencies []struct {
			Value string `json:"value"`
			Count int    `json:"count"`
		} `json:"frequencies"`
	}
	_ = json.Unmarshal(out, &got)
	if len(got.Frequencies) != 2 {
		t.Fatalf("%+v", got.Frequencies)
	}
	// Both count 1 → "" before "1" in UTF-8 order
	if got.Frequencies[0].Value != "" || got.Frequencies[0].Count != 1 {
		t.Fatalf("first %+v", got.Frequencies[0])
	}
	if got.Frequencies[1].Value != "1" || got.Frequencies[1].Count != 1 {
		t.Fatalf("second %+v", got.Frequencies[1])
	}
}

func TestFieldFrequency_rejectsTopNOver100(t *testing.T) {
	h := FieldFrequencyHandler(validate.Limits{})
	args, _ := json.Marshal(map[string]any{
		"path":        "x",
		"root":        t.TempDir(),
		"field_index": 0,
		"top_n":       101,
	})
	_, err := h(context.Background(), args)
	if err == nil || !strings.Contains(err.Error(), "top_n") {
		t.Fatalf("got %v", err)
	}
}
