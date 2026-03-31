package textscan

import "bytes"

// Normalize returns a copy of line with Unicode lowercasing when caseSensitive is false.
func Normalize(line []byte, caseSensitive bool) []byte {
	if caseSensitive {
		return bytes.Clone(line)
	}
	return bytes.ToLower(bytes.Clone(line))
}
