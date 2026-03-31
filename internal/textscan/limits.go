package textscan

// DefaultMaxLines caps how many newline-delimited lines ScanLines will process.
const DefaultMaxLines = 1_000_000

// DefaultMaxLineBytes is the default maximum size of one line (excluding the delimiter) for bufio.Scanner.
const DefaultMaxLineBytes = 4 << 20 // 4 MiB

// ScanLimits bounds streaming reads. Zero values are replaced by defaults in ScanLines.
type ScanLimits struct {
	MaxLines     int
	MaxLineBytes int
}

func (l ScanLimits) withDefaults() ScanLimits {
	out := l
	if out.MaxLines <= 0 {
		out.MaxLines = DefaultMaxLines
	}
	if out.MaxLineBytes <= 0 {
		out.MaxLineBytes = DefaultMaxLineBytes
	}
	return out
}
