package constraints

import (
	"testing"
	"unicode/utf8"
)

func TestParseSafeFilename(t *testing.T) {
	t.Run("ParseSafeFilename", func(t *testing.T) {
		valid := []string{"shakespeare", "hamlet-act-1", "faust-teil-2"}
		for _, v := range valid {
			if _, err := ParseSafeFilename(v); err != nil {
				t.Errorf("ParseSafeFilename(%q) unexpected error: %v", v, err)
			}
		}

		invalid := []string{
			"ab",             // minCodePoints
			"Upper-Case",     // only lower case
			"double--hyphen", // consecutive hyphens
			"-leading",       // leading hyphen
			"trailing-",      // trailing hyphen
			"con",            // Windows reserved
			"nul",            // Windows reserved
			"has space",      // whitespace
		}
		for _, inv := range invalid {
			if _, err := ParseSafeFilename(inv); err == nil {
				t.Errorf("ParseSafeFilename(%q) succeeded, expected error", inv)
			}
		}
	})
}

func TestFirstBytes(t *testing.T) {
	t.Run("FirstBytes not split UTF-8 rune", func(t *testing.T) {
		// "ä" = 2 bytes; "🦀" = 4 bytes
		input := "abc🦀def" // 'a'(1) 'b'(1) 'c'(1) '🦀'(4) = 7 bytes up to end of crab
		for maxBytes := 1; maxBytes <= len(input); maxBytes++ {
			got := FirstBytes(input, maxBytes)
			if len(got) > maxBytes {
				t.Errorf("FirstBytes(%q, %d) len=%d > %d", input, maxBytes, len(got), maxBytes)
			}
			if !utf8.ValidString(got) {
				t.Errorf("FirstBytes(%q, %d) produced invalid UTF-8: %x", input, maxBytes, got)
			}
		}
	})

}
