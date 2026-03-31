package textscan

import (
	"errors"
	"strings"
	"testing"
)

func TestScanLines_respectsLineCap(t *testing.T) {
	r := strings.NewReader("a\nb\nc\n")
	lim := ScanLimits{MaxLines: 2, MaxLineBytes: 1024}
	var seen int
	err := ScanLines(r, lim, func(line []byte, lineNo int) error {
		seen++
		if lineNo != seen {
			t.Fatalf("lineNo %d seen %d", lineNo, seen)
		}
		return nil
	})
	if !errors.Is(err, ErrTooManyLines) {
		t.Fatalf("got %v", err)
	}
	if seen != 2 {
		t.Fatalf("seen %d", seen)
	}
}

func TestScanLines_lineTooLong(t *testing.T) {
	long := strings.Repeat("x", 200)
	r := strings.NewReader(long + "\n")
	lim := ScanLimits{MaxLines: 10, MaxLineBytes: 100}
	err := ScanLines(r, lim, func([]byte, int) error { return nil })
	if !errors.Is(err, ErrLineTooLong) {
		t.Fatalf("got %v", err)
	}
}

func TestScanLines_defaultLimits(t *testing.T) {
	r := strings.NewReader("one\n")
	var got string
	err := ScanLines(r, ScanLimits{}, func(line []byte, lineNo int) error {
		got = string(line)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "one" {
		t.Fatalf("got %q", got)
	}
}
