package textscan

import (
	"errors"
	"strings"
	"testing"
)

func TestScanLines_lineTooLong(t *testing.T) {
	long := strings.Repeat("x", 200) + "\n"
	r := strings.NewReader(long)
	err := ScanLines(r, ScanLimits{MaxLines: 10, MaxLineBytes: 100}, func([]byte, int) error { return nil })
	if !errors.Is(err, ErrLineTooLong) {
		t.Fatalf("got %v", err)
	}
}

func TestScanLines_lineCap(t *testing.T) {
	r := strings.NewReader("a\nb\nc\n")
	var n int
	err := ScanLines(r, ScanLimits{MaxLines: 2, MaxLineBytes: 1024}, func([]byte, int) error {
		n++
		return nil
	})
	if !errors.Is(err, ErrTooManyLines) || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}
