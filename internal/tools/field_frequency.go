package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"wc-stats-mcp/internal/textscan"
	"wc-stats-mcp/internal/validate"
)

const (
	ffDefaultTopN = 20
	ffMaxTopN     = 100
)

// FieldFrequencyArgs matches unique_values plus ranking options.
type FieldFrequencyArgs struct {
	Path         string `json:"path"`
	Root         string `json:"root,omitempty"`
	MaxBytes     int64  `json:"max_bytes"`
	Delimiter    string `json:"delimiter"`
	FieldIndex   *int   `json:"field_index"`
	TrimSpace    *bool  `json:"trim_space"`
	MaxLines     int    `json:"max_lines"`
	TopN         int    `json:"top_n"`
	IncludeEmpty *bool  `json:"include_empty"`
}

// FieldFrequencyHandler counts occurrences of a delimited field value per line (not CSV-safe; same as unique_values).
func FieldFrequencyHandler(vlim validate.Limits) Handler {
	return func(ctx context.Context, arguments json.RawMessage) (json.RawMessage, error) {
		_ = ctx
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}
		dec := json.NewDecoder(bytes.NewReader(arguments))
		dec.DisallowUnknownFields()
		var args FieldFrequencyArgs
		if err := dec.Decode(&args); err != nil {
			return nil, fmt.Errorf("arguments: %w", err)
		}
		if args.Path == "" {
			return nil, fmt.Errorf("path is required")
		}
		if args.FieldIndex == nil {
			return nil, fmt.Errorf("field_index is required")
		}
		idx := *args.FieldIndex
		if idx < 0 || idx > uvMaxFieldIndex {
			return nil, fmt.Errorf("field_index must be between 0 and %d", uvMaxFieldIndex)
		}

		root := strings.TrimSpace(args.Root)
		if root == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("default root (cwd): %w", err)
			}
			root = cwd
		}

		delim := args.Delimiter
		if delim == "" {
			delim = uvDefaultDelimiter
		}
		if len(delim) > uvMaxDelimiterBytes {
			return nil, fmt.Errorf("delimiter must be at most %d bytes", uvMaxDelimiterBytes)
		}

		trim := true
		if args.TrimSpace != nil {
			trim = *args.TrimSpace
		}

		includeEmpty := false
		if args.IncludeEmpty != nil {
			includeEmpty = *args.IncludeEmpty
		}

		topN := args.TopN
		if topN <= 0 {
			topN = ffDefaultTopN
		}
		if topN > ffMaxTopN {
			return nil, fmt.Errorf("top_n must be at most %d", ffMaxTopN)
		}

		maxLines := args.MaxLines
		if maxLines <= 0 {
			maxLines = uvDefaultMaxLines
		}
		if maxLines > uvHardMaxLines {
			maxLines = uvHardMaxLines
		}

		reqLim := vlim.WithRequestMaxBytes(args.MaxBytes)
		resolved, _, err := validate.ResolveUnderRoot(root, args.Path, reqLim)
		if err != nil {
			return nil, fmt.Errorf("validate: %w", err)
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
		if err := textscan.ScanLines(f, slim, func(line []byte, _ int) error {
			val, ok := extractSplitField(line, delim, idx, trim)
			if !ok {
				return nil
			}
			if val == "" && !includeEmpty {
				return nil
			}
			counts[val]++
			return nil
		}); err != nil {
			return nil, err
		}

		type pair struct {
			Value string `json:"value"`
			Count int    `json:"count"`
		}
		pairs := make([]pair, 0, len(counts))
		for v, c := range counts {
			pairs = append(pairs, pair{Value: v, Count: c})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].Count != pairs[j].Count {
				return pairs[i].Count > pairs[j].Count
			}
			return pairs[i].Value < pairs[j].Value
		})
		if len(pairs) > topN {
			pairs = pairs[:topN]
		}

		out := struct {
			Path        string `json:"path"`
			FieldIndex  int    `json:"field_index"`
			TopN        int    `json:"top_n"`
			Frequencies []pair `json:"frequencies"`
		}{
			Path:        resolved,
			FieldIndex:  idx,
			TopN:        topN,
			Frequencies: pairs,
		}
		return json.Marshal(out)
	}
}
