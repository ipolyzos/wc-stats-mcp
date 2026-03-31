package validate

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUnderRoot_emptyPath(t *testing.T) {
	root := t.TempDir()
	_, _, err := ResolveUnderRoot(root, "   ", Limits{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

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

func TestResolveUnderRoot_traversalDotDotInUserPath(t *testing.T) {
	root := t.TempDir()
	_, _, err := ResolveUnderRoot(root, "../outside", Limits{})
	if !errors.Is(err, ErrDotDot) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveUnderRoot_nestedRelative_ok(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(sub, "f.txt")
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	res, _, err := ResolveUnderRoot(root, filepath.Join("sub", "f.txt"), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(res) != filepath.Clean(f) {
		t.Fatalf("resolved %q want %q", res, f)
	}
}

func TestResolveUnderRoot_absoluteUserPathRejected(t *testing.T) {
	root := t.TempDir()
	_, _, err := ResolveUnderRoot(root, "/etc/passwd", Limits{})
	if !errors.Is(err, ErrAbsolutePath) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveUnderRoot_notFound(t *testing.T) {
	root := t.TempDir()
	_, _, err := ResolveUnderRoot(root, "missing.txt", Limits{})
	if !errors.Is(err, ErrNotExist) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveUnderRoot_notRegularFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "d")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, err := ResolveUnderRoot(root, "d", Limits{})
	if !errors.Is(err, ErrNotRegular) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveUnderRoot_tooLarge(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "big.bin")
	if err := os.WriteFile(p, []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := ResolveUnderRoot(root, "big.bin", Limits{MaxFileBytes: 3})
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveUnderRoot_symlinkInsideRoot_ok(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "real.txt")
	if err := os.WriteFile(target, []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "via")
	if err := os.Symlink("real.txt", link); err != nil {
		t.Skip("symlinks not supported:", err)
	}
	res, fi, err := ResolveUnderRoot(root, "via", Limits{MaxFileBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	if fi.Size() != 2 {
		t.Fatalf("size %d", fi.Size())
	}
	if filepath.Clean(res) != filepath.Clean(target) {
		t.Fatalf("resolved %q want %q", res, target)
	}
}

func TestResolveUnderRoot_symlinkEscapesRoot_rejected(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	outFile := filepath.Join(out, "escape.txt")
	if err := os.WriteFile(outFile, []byte("nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "bad")
	if err := os.Symlink(outFile, link); err != nil {
		t.Skip("symlinks not supported:", err)
	}
	_, _, err := ResolveUnderRoot(root, "bad", Limits{})
	if !errors.Is(err, ErrTraversal) {
		t.Fatalf("expected ErrTraversal, got %v", err)
	}
}

func TestResolveUnderRoot_symlinkToDir_rejected(t *testing.T) {
	root := t.TempDir()
	d := filepath.Join(root, "dir")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "todir")
	if err := os.Symlink("dir", link); err != nil {
		t.Skip("symlinks not supported:", err)
	}
	_, _, err := ResolveUnderRoot(root, "todir", Limits{})
	if !errors.Is(err, ErrNotRegular) {
		t.Fatalf("got %v", err)
	}
}
