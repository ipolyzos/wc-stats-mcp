package textscan

import (
	"bufio"
	"errors"
	"io"

	"wc-stats-mcp/internal/validate"
)

// ScanLines reads newline-delimited lines; line numbers start at 1.
func ScanLines(r io.Reader, lim ScanLimits, fn func(line []byte, lineNo int) error) error {
	lim = lim.withDefaults()
	sc := bufio.NewScanner(r)
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

// ScanLinesFile opens path under root, then streams lines.
func ScanLinesFile(root, userPath string, vlim validate.Limits, slim ScanLimits, fn func(line []byte, lineNo int) error) error {
	f, _, err := OpenFile(root, userPath, vlim)
	if err != nil {
		return err
	}
	defer f.Close()
	return ScanLines(f, slim, fn)
}
