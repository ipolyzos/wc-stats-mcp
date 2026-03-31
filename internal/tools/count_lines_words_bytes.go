package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"wc-stats-mcp/internal/exec"
	"wc-stats-mcp/internal/validate"
)

// CountLinesWordsBytesArgs is the JSON shape for tools/call arguments.
type CountLinesWordsBytesArgs struct {
	Path     string `json:"path"`
	Root     string `json:"root,omitempty"`
	MaxBytes int64  `json:"max_bytes"`
}

// CountLinesWordsBytesHandler runs wc -l -w -c on a path validated under root.
func CountLinesWordsBytesHandler(r *exec.Runner, lim validate.Limits) Handler {
	if r == nil {
		r = exec.NewRunner()
	}
	return func(ctx context.Context, arguments json.RawMessage) (json.RawMessage, error) {
		_ = ctx
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}
		dec := json.NewDecoder(bytes.NewReader(arguments))
		dec.DisallowUnknownFields()
		var args CountLinesWordsBytesArgs
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

		reqLim := lim.WithRequestMaxBytes(args.MaxBytes)
		resolved, _, err := validate.ResolveUnderRoot(root, args.Path, reqLim)
		if err != nil {
			return nil, err
		}

		stdout, stderr, err := r.Run("wc", []string{"-l", "-w", "-c", resolved})
		es := strings.TrimSpace(string(stderr))
		if err != nil {
			if es != "" {
				return nil, fmt.Errorf("wc: %w (%s)", err, es)
			}
			return nil, fmt.Errorf("wc: %w", err)
		}

		lines, words, byteCount, perr := ParseWCCounts(stdout)
		if perr != nil {
			if es != "" {
				return nil, fmt.Errorf("%v (stderr: %s)", perr, es)
			}
			return nil, perr
		}

		out := struct {
			Path  string `json:"path"`
			Lines int    `json:"lines"`
			Words int    `json:"words"`
			Bytes int    `json:"bytes"`
		}{
			Path:  resolved,
			Lines: lines,
			Words: words,
			Bytes: byteCount,
		}
		return json.Marshal(out)
	}
}
