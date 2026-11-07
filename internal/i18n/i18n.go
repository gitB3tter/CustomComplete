package i18n

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type contextKey struct{}

var SupportedLanguages = []language.Tag{
	language.English,
	language.German,
}

// english is the default languague
var matcher = language.NewMatcher(SupportedLanguages)

// translation key
type Key string

var Translations = map[language.Tag]map[Key]string{
	language.English: translationsEn,
	language.German:  translationsDe,
}

// Takes the raw header string, parses it, and returns the best matching tag.
func ResolveFromHeader(acceptLangHeader string) language.Tag {
	tags, _, _ := language.ParseAcceptLanguage(acceptLangHeader)
	_, idx, _ := matcher.Match(tags...)
	return SupportedLanguages[idx]
}

func WithLangTag(ctx context.Context, langTag language.Tag) context.Context {
	return context.WithValue(ctx, contextKey{}, langTag)
}

func TranslateFuncFromContext(ctx context.Context) func(key Key, args ...any) string {
	var lang language.Tag
	var ok bool
	if lang, ok = ctx.Value(contextKey{}).(language.Tag); !ok {
		lang = language.English
	}
	return func(key Key, args ...any) string {
		return Translate(lang, key, args...)
	}
}

func Func(langTag language.Tag) func(key Key, args ...any) string {
	return func(key Key, args ...any) string {
		return Translate(langTag, key, args...)
	}
}

func LangTagFromContext(ctx context.Context) language.Tag {
	if lang, ok := ctx.Value(contextKey{}).(language.Tag); ok {
		return lang
	} else {
		return language.English
	}
}

func Translate(lang language.Tag, key Key, args ...any) string {
	var out string

	if dict, ok := Translations[lang]; ok {
		if val, ok := dict[key]; ok {
			out = val
		}
	}

	// english fallback
	if out == "" {
		if val, ok := Translations[language.English][key]; ok {
			out = val
		}
	}

	// key fallback
	if out == "" {
		out = string(key)
	}

	if len(args) > 0 && strings.Contains(out, "%") {
		return fmt.Sprintf(out, args...)
	}

	return out
}

// template-friendly translate func.
// first can be TranslationKey, string(TranslationKey), any(TranslationKey) or []string or []any with the former as first element
func TemplateTranslate(lang language.Tag) func(first any, rest ...any) string {
	return func(first any, rest ...any) string {
		var key Key
		var finalArgs []any

		switch v := first.(type) {
		case []string:
			if len(v) == 0 {
				return ""
			}
			key = Key(v[0])
			for _, s := range v[1:] {
				finalArgs = append(finalArgs, s)
			}
		case []any:
			if len(v) == 0 {
				return ""
			}
			key = Key(fmt.Sprint(v[0]))
			finalArgs = append(finalArgs, v[1:]...)
		case Key:
			key = v
		case string:
			key = Key(v)
		default:
			key = Key(fmt.Sprint(v))
		}

		for _, arg := range rest {
			switch slice := arg.(type) {
			case []string:
				for _, s := range slice {
					finalArgs = append(finalArgs, s)
				}
			case []any:
				finalArgs = append(finalArgs, slice...)
			default:
				finalArgs = append(finalArgs, arg)
			}
		}

		// 3. Call your real Translate function
		return Translate(lang, key, finalArgs...)
	}
}

func FormatBytesLocale(b int64, tag language.Tag) string {
	const (
		KB = 1000.0
		MB = 1000.0 * KB
	)

	p := message.NewPrinter(tag)

	switch {
	case b < 1000:
		return p.Sprintf("%d B", b)
	case b < 1000*1000:
		return p.Sprintf("%.1f kB", float64(b)/KB)
	default:
		return p.Sprintf("%.1f MB", float64(b)/MB)
	}
}
