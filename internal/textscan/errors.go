package textscan

import "errors"

var (
	// ErrLineTooLong is returned when a single line exceeds ScanLimits.MaxLineBytes (bufio.ErrTooLong).
	ErrLineTooLong = errors.New("textscan: line exceeds MaxLineBytes")

	// ErrTooManyLines is returned when more than ScanLimits.MaxLines newline-delimited records are read.
	ErrTooManyLines = errors.New("textscan: line count exceeds MaxLines")
)
