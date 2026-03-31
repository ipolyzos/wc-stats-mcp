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

func TestTopTerms_golden(t *testing.T) {
	root := t.TempDir()
	sample, err := os.ReadFile("testdata/top_terms_sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sample.txt"), sample, 0o600); err != nil {
		t.Fatal(err)
	}

	h := TopTermsHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, err := json.Marshal(map[string]any{
		"path":           "sample.txt",
		"root":           root,
		"top_n":          3,
		"min_len":        3,
		"case_sensitive": false,
		"max_lines":      1_000_000,
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
		TopN  int    `json:"top_n"`
		Terms []struct {
			Term  string `json:"term"`
			Count int    `json:"count"`
		} `json:"terms"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.TopN != 3 {
		t.Fatalf("top_n %d", got.TopN)
	}
	if len(got.Terms) != 3 {
		t.Fatalf("terms: %+v", got.Terms)
	}
	want := []struct {
		term  string
		count int
	}{
		{"apple", 3},
		{"banana", 2},
		{"cherry", 1},
	}
	for i := range want {
		if got.Terms[i].Term != want[i].term || got.Terms[i].Count != want[i].count {
			t.Fatalf("idx %d: got %+v want %+v", i, got.Terms[i], want[i])
		}
	}
}

func TestTopTerms_topNClampedDistinct(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "t.txt"), []byte("aaa bbb\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	h := TopTermsHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{"path": "t.txt", "root": root, "top_n": 10})
	raw, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Terms []struct {
			Term  string `json:"term"`
			Count int    `json:"count"`
		} `json:"terms"`
	}
	_ = json.Unmarshal(raw, &got)
	if len(got.Terms) != 2 {
		t.Fatalf("%+v", got.Terms)
	}
}

func TestTopTerms_rejectsTopNOver100(t *testing.T) {
	h := TopTermsHandler(validate.Limits{})
	args, _ := json.Marshal(map[string]any{"path": "x", "root": t.TempDir(), "top_n": 101})
	_, err := h(context.Background(), args)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTopTerms_offsetPagination(t *testing.T) {
	root := t.TempDir()
	content := "alpha beta gamma delta\nalpha beta gamma\nalpha beta\nalpha\n"
	if err := os.WriteFile(filepath.Join(root, "t.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	h := TopTermsHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{
		"path":      "t.txt",
		"root":      root,
		"top_n":     2,
		"min_len":   1,
		"offset":    1,
		"max_lines": 1_000_000,
	})
	raw, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Offset      int  `json:"offset"`
		TotalUnique int  `json:"total_unique"`
		HasMore     bool `json:"has_more"`
		NextOffset  int  `json:"next_offset"`
		Terms       []struct {
			Term  string `json:"term"`
			Count int    `json:"count"`
		} `json:"terms"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Offset != 1 || got.TotalUnique != 4 || !got.HasMore || got.NextOffset != 3 {
		t.Fatalf("pagination metadata %+v", got)
	}
	// Full sorted order is alpha(4), beta(3), gamma(2), delta(1); offset=1, top_n=2 => beta,gamma
	if len(got.Terms) != 2 || got.Terms[0].Term != "beta" || got.Terms[1].Term != "gamma" {
		t.Fatalf("terms %+v", got.Terms)
	}
}

func TestTopTerms_offsetTooLargeReturnsEmptyPage(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "t.txt"), []byte("alpha beta"), 0o600); err != nil {
		t.Fatal(err)
	}
	h := TopTermsHandler(validate.Limits{MaxFileBytes: 1 << 20})
	args, _ := json.Marshal(map[string]any{"path": "t.txt", "root": root, "offset": 999, "top_n": 10, "min_len": 1})
	raw, err := h(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Offset      int   `json:"offset"`
		TotalUnique int   `json:"total_unique"`
		HasMore     bool  `json:"has_more"`
		Terms       []any `json:"terms"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Offset != got.TotalUnique || got.HasMore || len(got.Terms) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestTopTerms_rejectsNegativeOffset(t *testing.T) {
	h := TopTermsHandler(validate.Limits{})
	args, _ := json.Marshal(map[string]any{"path": "x", "root": t.TempDir(), "offset": -1})
	_, err := h(context.Background(), args)
	if err == nil || !strings.Contains(err.Error(), "offset") {
		t.Fatalf("expected offset error, got %v", err)
	}
}
