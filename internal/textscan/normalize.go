package textscan

import "bytes"

// Normalize returns a copy of line with Unicode lowercasing applied when caseSensitive is false.
// When caseSensitive is true, the bytes are copied unchanged.
// Invalid UTF-8 sequences are mapped by bytes.ToLower according to Go’s rules (replacement semantics per rune step).
func Normalize(line []byte, caseSensitive bool) []byte {
	if caseSensitive {
		return bytes.Clone(line)
	}
	return bytes.ToLower(bytes.Clone(line))
}
