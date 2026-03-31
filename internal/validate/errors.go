package validate

import "errors"

var (
	ErrInvalidInput = errors.New("validate: invalid input")
	ErrAbsolutePath = errors.New("validate: absolute paths are not allowed")
	ErrDotDot       = errors.New("validate: path contains ..")
	ErrTraversal    = errors.New("validate: path escapes root boundary")
	ErrNotExist     = errors.New("validate: path does not exist")
	ErrNotRegular   = errors.New("validate: not a regular file")
	ErrFileTooLarge = errors.New("validate: file exceeds maximum size")
	ErrSymlink      = errors.New("validate: symlink resolution failed")
)
