package textscan

import "errors"

var (
	ErrLineTooLong  = errors.New("textscan: line exceeds MaxLineBytes")
	ErrTooManyLines = errors.New("textscan: line count exceeds MaxLines")
)
