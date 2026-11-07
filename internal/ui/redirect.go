package ui

import (
	"customcomplete/internal/i18n"
	"html/template"
	"net/url"

	"golang.org/x/text/language"
)

var (
	templateRedirect map[language.Tag]*template.Template
)

type redirectPageData struct {
	Base          BaseDataHTML
	Message       string
	Location      *url.URL
	LocationLabel string
}

func SeeOtherHTML(lang language.Tag, message string, location *url.URL) ([]byte, error) {
	T := i18n.Func(lang)
	return BufferHTML(templateRedirect[lang], redirectPageData{
		Base: BaseDataHTML{
			Title:           T(i18n.HTTPRedirectSeeOther),
			MetaDescription: T(i18n.HTTPRedirectMetaDescription),
			Lang:            lang,
			NoIndex:         true,
		},
		Message:       message,
		Location:      location,
		LocationLabel: T(i18n.HTTPRedirectAnchorLabel),
	})
}
