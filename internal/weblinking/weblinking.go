package weblinking

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"

	"golang.org/x/text/language"
)

func AllAlternates(path Path, supportedTypes ...mediatypes.Type) Links {
	var links Links
	for _, langTag := range i18n.SupportedLanguages {
		l := New(Alternate, path)
		l.Language = langTag
		links = append(links, l)
	}
	for _, mediaType := range supportedTypes {
		l := New(Alternate, path)
		l.Type = mediaType
		links = append(links, l)
	}
	return links
}

func Alternates(path Path, curLang language.Tag, curType mediatypes.Type, supportedTypes ...mediatypes.Type) Links {
	return AlternatesFor(New(Alternate, path), curLang, curType, supportedTypes...)
}

func AlternatesFor(base Link, curLang language.Tag, curType mediatypes.Type, supportedTypes ...mediatypes.Type) Links {
	base.rel = Alternate
	var links Links
	for _, langTag := range i18n.SupportedLanguages {
		if langTag != curLang {
			links = append(links, base.WithLang(langTag))
		}
	}
	for _, mediaType := range supportedTypes {
		if mediaType != curType {
			links = append(links, base.WithType(mediaType))
		}
	}
	return links
}

func RootLink(langTag language.Tag) Link {
	return New(Up, RootPath).WithTitle(i18n.Translate(langTag, i18n.NavigationHome))
}
