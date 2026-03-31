package validate

import "errors"

var (
	// ErrInvalidInput is returned for empty root or user path (after trim).
	ErrInvalidInput = errors.New("validate: invalid input")

	// ErrAbsolutePath means userPath must be relative to root.
	ErrAbsolutePath = errors.New("validate: absolute paths are not allowed")

	// ErrDotDot means userPath contains a ".." path element.
	ErrDotDot = errors.New("validate: path contains ..")

	// ErrTraversal means the path resolves outside root (including via symlinks).
	ErrTraversal = errors.New("validate: path escapes root boundary")

	// ErrNotExist means the path does not exist.
	ErrNotExist = errors.New("validate: path does not exist")

	// ErrNotRegular means the resolved target is not a regular file.
	ErrNotRegular = errors.New("validate: not a regular file")

	// ErrFileTooLarge means the file exceeds MaxFileBytes.
	ErrFileTooLarge = errors.New("validate: file exceeds maximum size")

	// ErrSymlink means symlink resolution failed (loop, permission, etc.).
	ErrSymlink = errors.New("validate: symlink resolution failed")
)
