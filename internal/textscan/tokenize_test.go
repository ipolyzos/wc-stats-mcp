package textscan

import (
	"bytes"
	"testing"
)

func TestForEachNonLetterDigitToken_basic(t *testing.T) {
	var toks [][]byte
	err := ForEachNonLetterDigitToken([]byte("foo, bar  42!"), func(tk []byte) error {
		toks = append(toks, append([]byte(nil), tk...))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(toks) != 3 || string(toks[0]) != "foo" || string(toks[1]) != "bar" || string(toks[2]) != "42" {
		t.Fatalf("got %q", toks)
	}
}

func TestForEachNonLetterDigitToken_underscoreDelimiter(t *testing.T) {
	var toks [][]byte
	_ = ForEachNonLetterDigitToken([]byte("snake_case"), func(tk []byte) error {
		toks = append(toks, append([]byte(nil), tk...))
		return nil
	})
	if len(toks) != 2 || string(toks[0]) != "snake" || string(toks[1]) != "case" {
		t.Fatalf("got %q", toks)
	}
}

func TestForEachNonLetterDigitToken_unicodeLetters(t *testing.T) {
	line := []byte("Δέλτα English123")
	var toks [][]byte
	_ = ForEachNonLetterDigitToken(line, func(tk []byte) error {
		toks = append(toks, append([]byte(nil), tk...))
		return nil
	})
	if len(toks) != 2 {
		t.Fatalf("got %d tokens %q", len(toks), toks)
	}
	if string(toks[1]) != "English123" {
		t.Fatalf("second token %q", toks[1])
	}
}

func TestForEachNonLetterDigitToken_invalidUTF8(t *testing.T) {
	line := []byte{'a', 0xff, 'b'}
	var toks [][]byte
	_ = ForEachNonLetterDigitToken(line, func(tk []byte) error {
		toks = append(toks, append([]byte(nil), tk...))
		return nil
	})
	if len(toks) != 2 || string(toks[0]) != "a" || string(toks[1]) != "b" {
		t.Fatalf("got %q", toks)
	}
}

func TestForEachNonLetterDigitTokenNormalized_fold(t *testing.T) {
	var toks [][]byte
	_ = ForEachNonLetterDigitTokenNormalized([]byte("Foo BAR"), false, func(tk []byte) error {
		toks = append(toks, append([]byte(nil), tk...))
		return nil
	})
	if len(toks) != 2 || string(toks[0]) != "foo" || string(toks[1]) != "bar" {
		t.Fatalf("got %q", toks)
	}
}

func TestForEachNonLetterDigitTokenNormalized_sensitive(t *testing.T) {
	var toks [][]byte
	_ = ForEachNonLetterDigitTokenNormalized([]byte("Foo"), true, func(tk []byte) error {
		toks = append(toks, append([]byte(nil), tk...))
		return nil
	})
	if len(toks) != 1 || string(toks[0]) != "Foo" {
		t.Fatalf("got %q", toks)
	}
}

func TestNormalize(t *testing.T) {
	if got := Normalize([]byte("Ab"), true); !bytes.Equal(got, []byte("Ab")) {
		t.Fatalf("%q", got)
	}
	if got := Normalize([]byte("Ab"), false); !bytes.Equal(got, []byte("ab")) {
		t.Fatalf("%q", got)
	}
}
