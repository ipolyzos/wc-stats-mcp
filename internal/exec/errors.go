package exec

import (
	"errors"
	"fmt"
)

// Kind classifies subprocess runner failures.
type Kind string

const (
	InvalidArgument Kind = "InvalidArgument"
	NotFound        Kind = "NotFound"
	TooLarge        Kind = "TooLarge"
	Timeout         Kind = "Timeout"
	ExecFailed      Kind = "ExecFailed"
)

// Error is returned by Runner for all failure categories.
type Error struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("exec.%s: %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

// KindOf returns (k, true) if err chains to *Error.
func KindOf(err error) (Kind, bool) {
	var ee *Error
	if errors.As(err, &ee) {
		return ee.Kind, true
	}
	return "", false
}
