package textscan

const DefaultMaxLines = 1_000_000

const DefaultMaxLineBytes = 4 << 20

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
