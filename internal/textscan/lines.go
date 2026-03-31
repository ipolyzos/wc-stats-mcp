package textscan

import (
	"bufio"
	"errors"
	"io"

	"wc-stats-mcp/internal/validate"
)

// ScanLines reads newline-delimited lines from r. Line numbers start at 1.
// If a line (without the trailing '\n') exceeds lim.MaxLineBytes, ScanLines returns ErrLineTooLong.
// After MaxLines lines have been delivered to fn, ScanLines returns ErrTooManyLines without calling fn for further lines.
func ScanLines(r io.Reader, lim ScanLimits, fn func(line []byte, lineNo int) error) error {
	lim = lim.withDefaults()
	sc := bufio.NewScanner(r)
	// nil buffer: max token size is exactly lim.MaxLineBytes (large initial cap would defeat small caps).
	sc.Buffer(nil, lim.MaxLineBytes)

	var lineNo int
	for sc.Scan() {
		lineNo++
		if lineNo > lim.MaxLines {
			return ErrTooManyLines
		}
		if err := fn(sc.Bytes(), lineNo); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return ErrLineTooLong
		}
		return err
	}
	return nil
}

// ScanLinesFile opens path under root using validate.ResolveUnderRoot, then ScanLines the content.
func ScanLinesFile(root, userPath string, vlim validate.Limits, slim ScanLimits, fn func(line []byte, lineNo int) error) error {
	f, _, err := OpenFile(root, userPath, vlim)
	if err != nil {
		return err
	}
	defer f.Close()
	return ScanLines(f, slim, fn)
}
