package textscan

import (
	"unicode"
	"unicode/utf8"
)

// ForEachNonLetterDigitToken splits line into tokens and invokes fn for each.
//
// Rule (deterministic, UTF-8):
//   - Walk the line as UTF-8 runes. Replacement: invalid UTF-8 is consumed one byte at a time and treated as a delimiter (token boundary).
//   - A token is a maximal contiguous subsequence of runes r where
//     unicode.IsLetter(r) || unicode.IsNumber(r) (Unicode letter or number categories).
//   - Any other rune (whitespace, punctuation, symbols, underscore, etc.) ends the current token and is skipped.
//   - Empty tokens are not emitted.
//
// Tokens are emitted as newly allocated UTF-8 byte slices (not sub-slices of line).
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

// ForEachNonLetterDigitTokenNormalized applies Normalize(line, caseSensitive) before tokenization
// so case folding uses bytes.ToLower on the whole line (UTF-8 aware per Go).
func ForEachNonLetterDigitTokenNormalized(line []byte, caseSensitive bool, fn func(token []byte) error) error {
	norm := Normalize(line, caseSensitive)
	return ForEachNonLetterDigitToken(norm, fn)
}
