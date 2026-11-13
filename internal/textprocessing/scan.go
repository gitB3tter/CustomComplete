package textprocessing

import (
	"bufio"
	"bytes"
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

func newWordScanner(data []byte) *bufio.Scanner {
	s := bufio.NewScanner(bytes.NewReader(data))
	s.Buffer(nil, max(len(data), bufio.MaxScanTokenSize))
	s.Split(ScanNormalizedWords)
	return s
}

func WordCount(data []byte) (count int) {
	s := newWordScanner(data)
	for s.Scan() {
		count++
	}
	return count
}

func WordCountMin(data []byte, limit int) bool {
	if limit <= 0 {
		return true
	}
	s := newWordScanner(data)
	count := 0
	for s.Scan() {
		count++
		if count >= limit {
			return true
		}
	}
	return false
}
