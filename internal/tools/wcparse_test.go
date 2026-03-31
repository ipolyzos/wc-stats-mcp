package tools

import "testing"

func TestParseWCCounts_gnuStyle(t *testing.T) {
	line := "      2       3      42 /tmp/x y.txt\n"
	l, w, b, err := ParseWCCounts([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if l != 2 || w != 3 || b != 42 {
		t.Fatalf("got %d %d %d", l, w, b)
	}
}

func TestParseWCCounts_multilineFirstWins(t *testing.T) {
	out := "\n  1  1  10 trailing name\nignored\n"
	l, w, b, err := ParseWCCounts([]byte(out))
	if err != nil {
		t.Fatal(err)
	}
	if l != 1 || w != 1 || b != 10 {
		t.Fatalf("got %d %d %d", l, w, b)
	}
}

func TestParseWCCounts_empty(t *testing.T) {
	_, _, _, err := ParseWCCounts([]byte("   \n"))
	if err == nil {
		t.Fatal("expected error")
	}
}
