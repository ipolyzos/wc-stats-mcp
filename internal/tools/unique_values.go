package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"wc-stats-mcp/internal/textscan"
	"wc-stats-mcp/internal/validate"
)

const (
	uvDefaultDelimiter  = ","
	uvMaxDelimiterBytes = 3
	uvMaxFieldIndex     = 1000
	uvDefaultMaxUnique  = 1000
	uvHardMaxUnique     = 10_000
	uvDefaultMaxLines   = 1_000_000
	uvHardMaxLines      = 1_000_000
)

// ErrUniqueValuesCap stops line scanning once max_unique distinct values were collected.
var errUniqueValuesCap = errors.New("unique_values: reached max_unique")

// UniqueValuesArgs is the JSON shape for tools/call arguments.
type UniqueValuesArgs struct {
	Path       string `json:"path"`
	Root       string `json:"root,omitempty"`
	MaxBytes   int64  `json:"max_bytes"`
	Delimiter  string `json:"delimiter"`
	FieldIndex *int   `json:"field_index"` // required; 0-based
	TrimSpace  *bool  `json:"trim_space"`  // default true when omitted
	MaxUnique  int    `json:"max_unique"`
	MaxLines   int    `json:"max_lines"`
}

// UniqueValuesHandler streams delimited lines and returns sorted distinct field values.
//
// Limitation: splitting is naive strings.Split — no CSV quoting, no escape rules. Delimiters inside quoted fields are not supported.
func UniqueValuesHandler(vlim validate.Limits) Handler {
	return func(ctx context.Context, arguments json.RawMessage) (json.RawMessage, error) {
		_ = ctx
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}
		dec := json.NewDecoder(bytes.NewReader(arguments))
		dec.DisallowUnknownFields()
		var args UniqueValuesArgs
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

		maxUnique := args.MaxUnique
		if maxUnique <= 0 {
			maxUnique = uvDefaultMaxUnique
		}
		if maxUnique > uvHardMaxUnique {
			maxUnique = uvHardMaxUnique
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
			return nil, err
		}

		f, err := os.Open(resolved)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		seen := make(map[string]struct{})
		slim := textscan.ScanLimits{
			MaxLines:     maxLines,
			MaxLineBytes: textscan.DefaultMaxLineBytes,
		}
		scanErr := textscan.ScanLines(f, slim, func(line []byte, _ int) error {
			val, ok := extractSplitField(line, delim, idx, trim)
			if !ok {
				return nil
			}
			if _, dup := seen[val]; dup {
				return nil
			}
			if len(seen) >= maxUnique {
				return errUniqueValuesCap
			}
			seen[val] = struct{}{}
			return nil
		})
		if scanErr != nil && !errors.Is(scanErr, errUniqueValuesCap) {
			return nil, scanErr
		}

		outVals := make([]string, 0, len(seen))
		for v := range seen {
			outVals = append(outVals, v)
		}
		sort.Strings(outVals)

		out := struct {
			Path       string   `json:"path"`
			FieldIndex int      `json:"field_index"`
			Values     []string `json:"values"`
		}{
			Path:       resolved,
			FieldIndex: idx,
			Values:     outVals,
		}
		return json.Marshal(out)
	}
}

// extractSplitField splits line by delim and returns field idx (0-based). Short rows return ok=false.
func extractSplitField(line []byte, delim string, idx int, trim bool) (string, bool) {
	s := string(line)
	parts := strings.Split(s, delim)
	if idx < 0 || idx >= len(parts) {
		return "", false
	}
	v := parts[idx]
	if trim {
		v = strings.TrimSpace(v)
	}
	return v, true
}
