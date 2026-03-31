package validate

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUnderRoot_happyFile(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "a.txt")
	if err := os.WriteFile(p, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	res, fi, err := ResolveUnderRoot(root, "a.txt", Limits{MaxFileBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	if fi.Size() != 2 {
		t.Fatalf("size %d", fi.Size())
	}
	if filepath.Clean(res) != filepath.Clean(p) {
		t.Fatalf("resolved %q want %q", res, p)
	}
}

func TestResolveUnderRoot_dotDotRejected(t *testing.T) {
	root := t.TempDir()
	_, _, err := ResolveUnderRoot(root, "../x", Limits{})
	if !errors.Is(err, ErrDotDot) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveUnderRoot_notFound(t *testing.T) {
	root := t.TempDir()
	_, _, err := ResolveUnderRoot(root, "nope", Limits{})
	if !errors.Is(err, ErrNotExist) {
		t.Fatalf("got %v", err)
	}
}
