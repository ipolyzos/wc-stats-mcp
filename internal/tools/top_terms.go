package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"wc-stats-mcp/internal/textscan"
	"wc-stats-mcp/internal/validate"
)

const (
	topTermsDefaultTopN     = 20
	topTermsMaxTopN         = 100
	topTermsDefaultMinLen   = 3
	topTermsMaxMinLen       = 30
	topTermsDefaultMaxLines = 1_000_000
	topTermsHardMaxLines    = 1_000_000
)

// TopTermsArgs is the JSON shape for tools/call arguments.
type TopTermsArgs struct {
	Path          string `json:"path"`
	Root          string `json:"root,omitempty"`
	MaxBytes      int64  `json:"max_bytes"`
	TopN          int    `json:"top_n"`
	Offset        int    `json:"offset"`
	MinLen        int    `json:"min_len"`
	CaseSensitive bool   `json:"case_sensitive"`
	MaxLines      int    `json:"max_lines"`
}

// TopTermsHandler streams tokens from a file and returns the top-N frequent terms.
func TopTermsHandler(vlim validate.Limits) Handler {
	return func(ctx context.Context, arguments json.RawMessage) (json.RawMessage, error) {
		_ = ctx
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}
		dec := json.NewDecoder(bytes.NewReader(arguments))
		dec.DisallowUnknownFields()
		var args TopTermsArgs
		if err := dec.Decode(&args); err != nil {
			return nil, fmt.Errorf("arguments: %w", err)
		}
		if args.Path == "" {
			return nil, fmt.Errorf("path is required")
		}

		root := strings.TrimSpace(args.Root)
		if root == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("default root (cwd): %w", err)
			}
			root = cwd
		}

		topN := args.TopN
		if topN <= 0 {
			topN = topTermsDefaultTopN
		}
		if topN > topTermsMaxTopN {
			return nil, fmt.Errorf("top_n must be at most %d", topTermsMaxTopN)
		}
		if args.Offset < 0 {
			return nil, fmt.Errorf("offset must be non-negative")
		}

		minLen := args.MinLen
		if minLen <= 0 {
			minLen = topTermsDefaultMinLen
		}
		if minLen > topTermsMaxMinLen {
			return nil, fmt.Errorf("min_len must be at most %d", topTermsMaxMinLen)
		}

		maxLines := args.MaxLines
		if maxLines <= 0 {
			maxLines = topTermsDefaultMaxLines
		}
		if maxLines > topTermsHardMaxLines {
			maxLines = topTermsHardMaxLines
		}

		reqLim := vlim.WithRequestMaxBytes(args.MaxBytes)
		resolved, _, err := validate.ResolveUnderRoot(root, args.Path, reqLim)
		if err != nil {
			return nil, err
		}

		f, err := os.Open(resolved)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		counts := make(map[string]int)
		slim := textscan.ScanLimits{
			MaxLines:     maxLines,
			MaxLineBytes: textscan.DefaultMaxLineBytes,
		}
		err = textscan.ScanLines(f, slim, func(line []byte, _ int) error {
			return textscan.ForEachNonLetterDigitTokenNormalized(line, args.CaseSensitive, func(tok []byte) error {
				if utf8.RuneCountInString(string(tok)) < minLen {
					return nil
				}
				counts[string(tok)]++
				return nil
			})
		})
		if err != nil {
			return nil, err
		}

		type termCount struct {
			Term  string `json:"term"`
			Count int    `json:"count"`
		}
		pairs := make([]termCount, 0, len(counts))
		for term, c := range counts {
			pairs = append(pairs, termCount{Term: term, Count: c})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].Count != pairs[j].Count {
				return pairs[i].Count > pairs[j].Count
			}
			return pairs[i].Term < pairs[j].Term
		})
		totalUnique := len(pairs)
		start := args.Offset
		if start > totalUnique {
			start = totalUnique
		}
		end := start + topN
		if end > totalUnique {
			end = totalUnique
		}
		page := pairs[start:end]
		hasMore := end < totalUnique

		out := struct {
			Path        string      `json:"path"`
			TopN        int         `json:"top_n"`
			Offset      int         `json:"offset"`
			TotalUnique int         `json:"total_unique"`
			HasMore     bool        `json:"has_more"`
			NextOffset  int         `json:"next_offset,omitempty"`
			Terms       []termCount `json:"terms"`
		}{
			Path:        resolved,
			TopN:        topN,
			Offset:      start,
			TotalUnique: totalUnique,
			HasMore:     hasMore,
			Terms:       page,
		}
		if hasMore {
			out.NextOffset = end
		}
		return json.Marshal(out)
	}
}
