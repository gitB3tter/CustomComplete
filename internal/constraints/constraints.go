package constraints

import (
	"customcomplete/internal/errs"
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// not empty
type NonEmptyString string

// not empty when whitespace trimmed
type NonBlankString string

// useable as identifier in lists
type ShortString NonBlankString

// pattern `^[a-z0-9]+(-[a-z0-9]+)*$`, min. 5 codepoints, max 50 bytes, no reserved (windows) names
type SafeFilename NonBlankString

const (
	ShortStringMaxCodePoints  = 100
	safeFileNameMinCodePoints = 5
	safeFileNameMaxBytes      = 50
	safeFileNameFallback      = "untitled"
	SafeFilenamePattern       = `^[a-z0-9]+(-[a-z0-9]+)*$`
)

var SafeFilenameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// reserved names (windows)
var filenameReserved = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"com5": true, "com6": true, "com7": true, "com8": true,
	"com9": true, "lpt1": true, "lpt2": true, "lpt3": true,
	"lpt4": true, "lpt5": true, "lpt6": true, "lpt7": true,
	"lpt8": true, "lpt9": true,
}

// Unicode Code Points/Runes, js [...s].length
func codePoints(s string) int {
	return utf8.RuneCountInString(s)
}

func ParseNonEmptyString(s string) (NonEmptyString, error) {
	if s == "" {
		return "", errs.Required()
	}
	return NonEmptyString(s), nil
}

func ParseNonBlankString(s string) (NonBlankString, *errs.CodedError) {
	if strings.TrimSpace(s) == "" {
		return "", errs.Blank()
	}
	return NonBlankString(s), nil
}

func ParseShortString(s string) (ShortString, *errs.CodedError) {
	nonemptyStr, err := ParseNonBlankString(s)
	if err != nil {
		return "", err
	}
	length := codePoints(string(nonemptyStr))
	if length > ShortStringMaxCodePoints {
		return "", errs.MaxCodePoints(ShortStringMaxCodePoints, length)
	}
	return ShortString(nonemptyStr), nil
}

func ParseSafeFilename(str string) (SafeFilename, error) {
	safeFileName, err := parseSafeFilename(str)
	if err != nil {
		if coded, ok := errors.AsType[*errs.CodedError](err); ok && coded.Code() == errs.CodePattern {
			if canonical, err := parseSafeFilename(canonicalise(str)); err == nil {
				return "", errs.PatternWithCanonical(SafeFilenamePattern, string(canonical))
			}
		}
		return "", err
	}
	return safeFileName, nil
}

func parseSafeFilename(str string) (SafeFilename, error) {

	if n := codePoints(str); n < safeFileNameMinCodePoints {
		return "", errs.MinCodePoints(safeFileNameMinCodePoints, n)
	}

	if !utf8.ValidString(str) {
		return "", errs.InvalidUTF8()
	}

	if len(str) > safeFileNameMaxBytes {
		return "", errs.MaxBytes(safeFileNameMaxBytes)
	}

	if !SafeFilenameRegex.MatchString(str) {
		return "", errs.Pattern(SafeFilenamePattern)
	}

	if filenameReserved[str] {
		return "", errs.Invalid(errs.Params{errs.ParamRule: errs.RuleReserved})
	}

	if n := codePoints(str); n < safeFileNameMinCodePoints {
		return "", errs.MinCodePoints(safeFileNameMinCodePoints, n)
	}

	return SafeFilename(str), nil
}

func canonicalise(str string) string {
	normalized := strings.ToLower(norm.NFKC.String(strings.TrimSpace(str)))
	var sb strings.Builder
	hyphen := false
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			hyphen = false
		} else if !hyphen {
			sb.WriteRune('-')
			hyphen = true
		}
	}
	return strings.Trim(sb.String(), "-")
}

func CanonicalSafeFilename(s string) (SafeFilename, error) {
	return ParseSafeFilename(canonicalise(s))
}

// tries to parse a safe filename from the first 50 bytes of the first string on fail falling back to the next until ultimate fallback
// strings need to be <= 50 bytes (len(str))
func DeriveSafeFilename(candidates ...string) SafeFilename {

	for _, candidate := range candidates {
		if filename, err := CanonicalSafeFilename(FirstBytes(candidate, safeFileNameMaxBytes)); err == nil {
			return filename
		}
	}

	return SafeFilename(safeFileNameFallback)
}

func WordCount(s string) (count int) {
	inWord := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
			continue
		}
		if !inWord {
			count++
			inWord = true
		}
	}
	return count
}

func WordCountMin(s string, limit int) (ok bool) {
	inWord := false
	count := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
			continue
		}
		if !inWord {
			count++
			if count >= limit {
				return true
			}
			inWord = true
		}
	}
	return false
}

func FirstBytes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	end := limit
	for end > 0 && !utf8.RuneStart(s[end]) {
		end--
	}
	return strings.Clone(s[:end])
}
