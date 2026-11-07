package httpsyntax

import (
	"iter"
	"strings"
	"unicode/utf8"
)

const (
	HTTP_OWS = " \t" // SP / HTAB (RFC 9110 §5.6.3 Whitespace)
	//means unacceptable
	WeightZero Weight = 0
)

// yields each segment of field split by seperator, ignoring appereances in quotes, with the offset it starts at
// e.g. ',' for spliting a header into values/segments, ';' for splitting value and params
// see RFC 9110 §5.6.1 Lists (#rule ABNF Extension)/§12.5.1 Accept
func SplitList(field string, separator byte) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		start := 0
		inQuote, escaped := false, false

		// constrained to US-ASCII => iterate through bytes (not runes) (RFC 9110 §5.5)
		for i := 0; i < len(field); i++ {
			char := field[i]

			if inQuote {
				if escaped {
					escaped = false
					continue
				}
				if char == '\\' {
					escaped = true
					continue
				}
				if char == '"' {
					inQuote = false
					continue
				}
			}

			if char == '"' {
				inQuote = true
				continue
			}

			if char == separator {
				if segment := strings.Trim(field[start:i], HTTP_OWS); segment != "" && !yield(start, segment) {
					return
				}
				start = i + 1
			}
		}
		// last remaining segment
		if !inQuote {
			if segment := strings.Trim(field[start:], HTTP_OWS); segment != "" {
				yield(start, segment)
			}
		}
	}

}

// parsed HTTP q-value scaled to 0-1000 see RFC 9110 §12.4.2. Quality Values
type Weight int

// parses a Weight parameter ("q" value) from s
// RFC 9110
func ParseWeight(s string) (Weight, bool) {
	//invalid
	if s == "" || (s[0] != '0' && s[0] != '1') || len(s) > 5 || (len(s) > 1 && s[1] != '.') {
		return 0, false
	}
	//1
	if s == "1" || s == "1." || s == "1.0" || s == "1.00" || s == "1.000" {
		return 1000, true
	}
	//0
	if s == "0" || s == "0." {
		return 0, true
	}

	weight := 0
	multiplier := 100
	for i := 2; i < len(s); i++ {
		char := s[i]
		if char < '0' || char > '9' || (s[0] == '1' && char > '0') {
			return 0, false
		}
		weight += int(char-'0') * multiplier
		multiplier /= 10
	}
	return Weight(weight), true
}

// RFC 8187 §3.2.1
func isAttrChar(c byte) bool {
	return c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c >= '0' && c <= '9' ||
		strings.IndexByte("!#$&+-.^_`|~", c) >= 0
}

// RFC 3986 §2.1
// used for example to encode title* param of a Link (RFC 8288 §3.4.1)
// encodes UTF-8
func PercentEncode(s string) string {
	const hex = "0123456789ABCDEF"

	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isAttrChar(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f]) // last 4 bit
		}
	}
	return b.String()
}

// quote-escapes ASCII
func QuoteEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	return b.String()
}

func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
