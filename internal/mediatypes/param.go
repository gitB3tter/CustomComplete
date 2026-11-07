package mediatypes

import (
	"customcomplete/internal/httpsyntax"
	"iter"
	"mime"
	"strings"
)

func init() {
	if mime.FormatMediaType("a/b", map[string]string{"k": "v"}) != "a/b; k=v" {
		panic("Go stdlib changed relied upon mime.FormatMediaType spacing!")
	}
}

const (
	goMimeSep            = "; " // change if mime.FormatMediaType changes the delimiter between mediaType and/or params
	placeHolderMediaType = "a/b"
)

// valid media type parameters string, ordered alphabetically, charset name lowercased, values quoted where necessary
// '; ' (OWS after the ';') as separator because of mime.ParseMediaType doing that
type parameters string

var paramWithCaseInsensitiveValue = map[string]bool{
	"charset": true, // RFC 2046 §4.1.2
}

// requires valid params (like you get from mime.ParseMediaType) (e.g. no token-only)
func normalizeParams(params map[string]string) map[string]string {
	out := make(map[string]string, len(params))
	for name, value := range params {
		if paramWithCaseInsensitiveValue[name] {
			value = strings.ToLower(value)
		}
		out[name] = value
	}
	return out
}

// requires lowercased parameter-names (like you get from mime.ParseMediaType)
func canonizeParams(params map[string]string) parameters {
	formatted := mime.FormatMediaType(placeHolderMediaType, normalizeParams(params))
	lenBeforeFirstParam := len(placeHolderMediaType) + len(goMimeSep)
	if len(formatted) > lenBeforeFirstParam {
		return parameters(formatted[lenBeforeFirstParam:])
	}
	return ""
}

// gets the unquoted, unescaped value for the given parameter-name, if present
func (p parameters) Get(targetName string) (string, bool) {
	for name, value := range p.All() {
		if name == targetName {
			return value, true
		}
		if name > targetName {
			return "", false
		}
	}
	return "", false
}

// yields all keys and unquoted, unescaped values in alphabetical order
func (p parameters) All() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, segment := range httpsyntax.SplitList(string(p), ';') {

			name, value, ok := strings.Cut(segment, "=")
			if !ok {
				continue
			}
			if !yield(name, cleanValue(value)) {
				return
			}
		}
	}
}

// unqotes, unescapes a parameter-value
func cleanValue(raw string) string {
	raw = strings.Trim(raw, httpsyntax.HTTP_OWS)
	if len(raw) < 2 || raw[0] != '"' {
		//value not quoted
		return raw
	}
	//quoted => unquote and unescape
	//omit leading and trailing quotes
	inner := raw[1 : len(raw)-1]
	if !strings.Contains(inner, `\`) {
		//quoted but not escaping anything
		return inner
	}
	var b strings.Builder
	b.Grow(len(inner))
	escaped := false
	for i := 0; i < len(inner); i++ {
		char := inner[i]
		if !escaped && char == '\\' {
			escaped = true
		} else {
			escaped = false
			b.WriteByte(char)
		}
	}
	return b.String()
}
