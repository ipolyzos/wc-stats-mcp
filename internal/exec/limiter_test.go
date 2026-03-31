package exec

import (
	"bytes"
	"testing"
)

func TestCappedWriter_withinLimit(t *testing.T) {
	w := &CappedWriter{Stream: "stdout", Limit: 10}
	n, err := w.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	n, err = w.Write([]byte("world"))
	if err != nil || n != 5 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if string(w.Bytes()) != "helloworld" {
		t.Fatalf("got %q", w.Bytes())
	}
}

func TestCappedWriter_truncatesThenTooLarge(t *testing.T) {
	w := &CappedWriter{Stream: "stderr", Limit: 5}
	n, err := w.Write([]byte("abcdefgh"))
	if err == nil {
		t.Fatal("expected TooLarge")
	}
	k, ok := KindOf(err)
	if !ok || k != TooLarge {
		t.Fatalf("kind=%v ok=%v err=%v", k, ok, err)
	}
	if n != 5 {
		t.Fatalf("expected partial write 5, got %d", n)
	}
	if !bytes.Equal(w.Bytes(), []byte("abcde")) {
		t.Fatalf("buf=%q", w.Bytes())
	}
}

func TestCappedWriter_unlimitedWhenLimitNonPositive(t *testing.T) {
	w := &CappedWriter{Stream: "stdout", Limit: 0}
	payload := bytes.Repeat([]byte("z"), 5000)
	n, err := w.Write(payload)
	if err != nil || n != len(payload) {
		t.Fatalf("n=%d err=%v", n, err)
	}
}
