package validate

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DefaultMaxFileBytes is the default cap for ResolveUnderRoot (25 MiB).
const DefaultMaxFileBytes = 25 << 20

// Limits configures optional checks. Zero MaxFileBytes uses DefaultMaxFileBytes.
type Limits struct {
	MaxFileBytes int64
}

func (l Limits) maxBytes() int64 {
	if l.MaxFileBytes <= 0 {
		return DefaultMaxFileBytes
	}
	return l.MaxFileBytes
}

// WithRequestMaxBytes returns limits for this request: if requestMaxBytes > 0 and
// smaller than the effective server cap (l.maxBytes()), the returned Limits use
// requestMaxBytes as MaxFileBytes. Clients cannot raise the cap above the
// server's effective maximum.
func (l Limits) WithRequestMaxBytes(requestMaxBytes int64) Limits {
	if requestMaxBytes <= 0 {
		return l
	}
	effective := l.maxBytes()
	if requestMaxBytes < effective {
		out := l
		out.MaxFileBytes = requestMaxBytes
		return out
	}
	return l
}

// ResolveUnderRoot joins userPath under root, resolves symlinks to the final target,
// and verifies the target exists, is a regular file, stays within root, and respects MaxFileBytes.
//
// Symlink policy: intermediate and final symlinks are followed via filepath.EvalSymlinks.
// After resolution, the real path must still lie under the real root directory; links that
// escape the root are rejected (ErrTraversal).
//
// userPath must be relative and must not contain a ".." path element (ErrDotDot).
func ResolveUnderRoot(root, userPath string, lim Limits) (resolved string, info fs.FileInfo, err error) {
	root = strings.TrimSpace(root)
	userPath = strings.TrimSpace(userPath)
	if root == "" || userPath == "" {
		return "", nil, ErrInvalidInput
	}
	if filepath.IsAbs(userPath) {
		return "", nil, ErrAbsolutePath
	}
	if hasDotDotElement(userPath) {
		return "", nil, ErrDotDot
	}

	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", nil, err
	}
	rootEval, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", nil, err
	}
	rootEval, err = filepath.Abs(rootEval)
	if err != nil {
		return "", nil, err
	}

	candidate := filepath.Join(rootEval, filepath.FromSlash(userPath))
	candidate = filepath.Clean(candidate)

	if err := mustBeUnderRoot(rootEval, candidate); err != nil {
		return "", nil, err
	}

	if _, err := os.Lstat(candidate); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, ErrNotExist
		}
		return "", nil, err
	}

	resolved, err = filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", nil, errors.Join(ErrSymlink, err)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", nil, err
	}

	if err := mustBeUnderRoot(rootEval, resolved); err != nil {
		return "", nil, err
	}

	fi, err := os.Stat(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, ErrNotExist
		}
		return "", nil, err
	}
	if !fi.Mode().IsRegular() {
		return "", fi, ErrNotRegular
	}
	if fi.Size() > lim.maxBytes() {
		return "", fi, ErrFileTooLarge
	}
	return resolved, fi, nil
}

func hasDotDotElement(path string) bool {
	p := filepath.ToSlash(path)
	for _, part := range strings.Split(p, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

func mustBeUnderRoot(rootEval, target string) error {
	rel, err := filepath.Rel(rootEval, target)
	if err != nil {
		return ErrTraversal
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ErrTraversal
	}
	return nil
}
