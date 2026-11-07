package weblinking

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"net/url"

	"golang.org/x/text/language"
)

func AllAlternates(target *url.URL, supportedTypes ...mediatypes.Type) Links {
	var links Links
	for _, langTag := range i18n.SupportedLanguages {
		l := New(Alternate, target)
		l.Language = langTag
		links = append(links, l)
	}
	for _, mediaType := range supportedTypes {
		l := New(Alternate, target)
		l.Type = mediaType
		links = append(links, l)
	}
	return links
}

func Alternates(target *url.URL, curLang language.Tag, curType mediatypes.Type, supportedTypes ...mediatypes.Type) Links {
	var links Links
	alternate := New(Alternate, target)
	for _, langTag := range i18n.SupportedLanguages {
		if langTag != curLang {
			links = append(links, alternate.WithLang(langTag))
		}
	}
	for _, mediaType := range supportedTypes {
		if mediaType != curType {
			links = append(links, alternate.WithType(mediaType))
		}
	}
	return links
}
