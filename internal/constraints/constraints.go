package constraints

import (
	"bytes"
	"crypto/sha256"
	"customcomplete/internal/errs"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// not empty
type NonEmptyString string

// not empty when whitespace trimmed
type NotBlank string

// useable as identifier in lists
type Label NotBlank

// pattern `^[a-z0-9]+(-[a-z0-9]+)*$`,
// min. 5 codepoints (benefit: no reserved windows names (since all len < 5)),
// max 50 bytes (=> codepoints since only ascii)
type SafeFilename NotBlank

const (
	ShortStringMaxCodePoints        = 100
	SafeFileNameMinCodePoints       = 5
	SafeFileNameMaxBytes            = 50
	safeFileNameFallback            = "untitled"
	SafeFilenamePattern             = `^[a-z0-9]+(-[a-z0-9]+)*$`
	MaxFileSizeMB             int64 = 10
	MaxFileSize               int64 = MaxFileSizeMB * 1_000_000
)

var SafeFilenameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Unicode Code Points/Runes, js [...s].length
func codePoints(s string) int {
	return utf8.RuneCountInString(s)
}

func ParseNonEmptyString(s string) (NonEmptyString, error) {
	if s == "" {
		return "", new(errs.Required())
	}
	return NonEmptyString(s), nil
}

func ParseNonBlankString(s string) (NotBlank, error) {
	if strings.TrimSpace(s) == "" {
		return "", new(errs.Blank())
	}
	return NotBlank(s), nil
}

func ParseLabel(s string) (Label, error) {
	nonemptyStr, err := ParseNonBlankString(s)
	if err != nil {
		return "", err
	}
	length := codePoints(string(nonemptyStr))
	if length > ShortStringMaxCodePoints {
		return "", new(errs.MaxCodePoints(ShortStringMaxCodePoints, length))
	}
	return Label(nonemptyStr), nil
}

func ParseSafeFilename(str string) (SafeFilename, error) {
	safeFileName, err := parseSafeFilename(str)
	if err != nil {
		if errs.HasCode(err, errs.CodePattern) {
			if canonical, err := parseSafeFilename(canonicalise(str)); err == nil {
				return "", new(errs.PatternWithCanonical(SafeFilenamePattern, string(canonical)))
			}
		}
		return "", err
	}
	return safeFileName, nil
}

func parseSafeFilename(str string) (SafeFilename, error) {

	if n := codePoints(str); n < SafeFileNameMinCodePoints {
		return "", new(errs.MinCodePoints(SafeFileNameMinCodePoints, n))
	}

	if !utf8.ValidString(str) {
		return "", new(errs.InvalidUTF8())
	}

	if len(str) > SafeFileNameMaxBytes {
		return "", new(errs.MaxBytes(SafeFileNameMaxBytes))
	}

	if !SafeFilenameRegex.MatchString(str) {
		return "", new(errs.Pattern(SafeFilenamePattern))
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
		if filename, err := CanonicalSafeFilename(FirstBytes(candidate, SafeFileNameMaxBytes)); err == nil {
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

type Sha256Hex string

func NewSha256Hex(contents []byte) Sha256Hex {
	//dont destroy/corrupt the passed in contents
	trimmed := bytes.TrimSpace(contents)
	workingCopy := make([]byte, len(trimmed))
	copy(workingCopy, trimmed)
	hashArray := sha256.Sum256(collapseSpace(norm.NFC.Bytes(workingCopy)))
	return Sha256Hex(hex.EncodeToString(hashArray[:]))
}

// Collapses consecutive whitespace (spaces, tabs, newlines) into a single space in-place.
func collapseSpace(b []byte) []byte {
	n := 0
	lastWasSpace := false
	for _, v := range b {
		isSpace := v == ' ' || v == '\n' || v == '\t' || v == '\r'
		if isSpace {
			if !lastWasSpace {
				b[n] = ' '
				n++
			}
			lastWasSpace = true
		} else {
			b[n] = v
			n++
			lastWasSpace = false
		}
	}
	return b[:n]
}

type FileContents []byte

func ParseFileContents(contents []byte) (FileContents, error) {
	if len(contents) > int(MaxFileSize) {
		return nil, new(errs.MaxBytes(MaxFileSize))
	}
	return FileContents(contents), nil
}
