package validate

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const DefaultMaxFileBytes = 25 << 20

type Limits struct {
	MaxFileBytes int64
}

func (l Limits) maxBytes() int64 {
	if l.MaxFileBytes <= 0 {
		return DefaultMaxFileBytes
	}
	return l.MaxFileBytes
}

// ResolveUnderRoot joins userPath under root, resolves symlinks, and verifies a regular file under root within size limit.
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
