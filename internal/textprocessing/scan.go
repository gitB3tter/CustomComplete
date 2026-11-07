package textprocessing

import (
	"unicode"
	"unicode/utf8"
)

// allowed between two wordChar
func isWordConnector(r rune) bool {
	switch r {
	case '\'', '’', '-', '‑':
		return true
	}
	return false
}

// allowed between two numbers
func isNumberConnector(r rune) bool {
	switch r {
	case '.', ',':
		return true
	}
	return false
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// bufio.SplitFunc stripping leading and trailing punctuation, keeping internal hyphens and apostrophes and splitting on em-dashes
func ScanNormalizedWords(data []byte, atEOF bool) (advance int, token []byte, err error) {

	start := 0
	for width := 0; start < len(data); start += width {
		r, w := utf8.DecodeRune(data[start:])
		if isWordChar(r) {
			break
		}
		width = w
	}

	if start == len(data) {
		if atEOF {
			return len(data), nil, nil
		}
		return start, nil, nil
	}

	var prevRune rune
	for i := start; i < len(data); {

		if !utf8.FullRune(data[i:]) && !atEOF {
			return start, nil, nil
		}

		r, width := utf8.DecodeRune(data[i:])

		if isWordChar(r) {
			prevRune = r
			i += width
			continue
		}

		if isWordConnector(r) || (isNumberConnector(r) && unicode.IsNumber(prevRune)) {
			if !utf8.FullRune(data[i+width:]) {
				if !atEOF {
					return start, nil, nil
				}
				return i + width, data[start:i], nil
			}

			rNext, nextWidth := utf8.DecodeRune(data[i+width:])
			if (isWordConnector(r) && isWordChar(rNext)) ||
				(isNumberConnector(r) && unicode.IsNumber(rNext)) {
				prevRune = rNext
				i += width + nextWidth
				continue
			}
		}

		return i + width, data[start:i], nil
	}

	if atEOF {
		return len(data), data[start:], nil
	}

	return start, nil, nil
}
