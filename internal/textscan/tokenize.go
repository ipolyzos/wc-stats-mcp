package textscan

import (
	"unicode"
	"unicode/utf8"
)

// ForEachNonLetterDigitToken walks UTF-8 text and emits maximal runs of runes where
// unicode.IsLetter(r) || unicode.IsNumber(r). Any other rune splits tokens; invalid UTF-8
// bytes are treated as delimiters. Tokens are allocated copies (not sub-slices of line).
func ForEachNonLetterDigitToken(line []byte, fn func(token []byte) error) error {
	var cur []byte
	flush := func() error {
		if len(cur) == 0 {
			return nil
		}
		tok := make([]byte, len(cur))
		copy(tok, cur)
		cur = cur[:0]
		return fn(tok)
	}

	for len(line) > 0 {
		r, sz := utf8.DecodeRune(line)
		if sz == 0 {
			break
		}
		line = line[sz:]
		if r == utf8.RuneError && sz == 1 {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			cur = utf8.AppendRune(cur, r)
		} else {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	return flush()
}

// ForEachNonLetterDigitTokenNormalized applies Normalize(line, caseSensitive) before tokenization.
func ForEachNonLetterDigitTokenNormalized(line []byte, caseSensitive bool, fn func(token []byte) error) error {
	return ForEachNonLetterDigitToken(Normalize(line, caseSensitive), fn)
}
