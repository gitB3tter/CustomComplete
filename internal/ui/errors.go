package ui

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/weblinking"
	"html/template"
	"strings"

	"golang.org/x/text/language"
)

var (
	templateErr map[language.Tag]*template.Template
)

type errorPageData struct {
	Base    BaseDataHTML
	Message string
}

func renderError(lang language.Tag, back weblinking.Link, title, metaDesc, message string) ([]byte, error) {
	return BufferHTML(templateErr[lang], errorPageData{
		Base: BaseDataHTML{
			Title:           title,
			MetaDescription: metaDesc,
			Lang:            lang,
			NoIndex:         true,
			HasError:        true,
			Back:            back,
		},
		Message: message,
	})
}

func InternalServerErrorHTML(lang language.Tag, back weblinking.Link) ([]byte, error) {
	T := i18n.Func(lang)
	msg := T(i18n.HTTPErrorInternalServerErrorTryAgain)
	return renderError(lang, back, T(i18n.CommonErrorService), msg, msg)
}
func RequestBodyNotAllowedHTML(lang language.Tag, back weblinking.Link) ([]byte, error) {
	T := i18n.Func(lang)
	msg := T(i18n.HTTPErrorRequestBodyNotAllowed)
	return renderError(lang, back, T(i18n.CommonErrorSummaryTitle), msg, msg)
}
func NotFoundHTML(lang language.Tag, back weblinking.Link) ([]byte, error) {
	T := i18n.Func(lang)
	msg := T(i18n.HTTPErrorNotFoundDescription)
	return renderError(lang, back, T(i18n.HTTPErrorNotFound), msg, msg)
}
func ContentTooLargeHTML(lang language.Tag, limitFmt string, back weblinking.Link) ([]byte, error) {
	T := i18n.Func(lang)
	msg := T(i18n.HTTPErrorContentTooLargeDescription, limitFmt)
	return renderError(lang, back, T(i18n.CommonErrorSummaryTitle), msg, msg)
}
func UnsupportedMediaTypeHTML(lang language.Tag, back weblinking.Link, provided mediatypes.Type, supported mediatypes.Types) ([]byte, error) {
	T := i18n.Func(lang)
	title := T(i18n.HTTPErrorUnsupportedMediaType)
	meta := T(i18n.HTTPErrorUnsupportedMediaTypeMetaDescription)
	msg := T(i18n.HTTPErrorUnsupportedMediaTypeDescription, provided, supported)
	return renderError(lang, back, title, meta, msg)
}
func BadRequestHTML(lang language.Tag, back weblinking.Link, detail string) ([]byte, error) {
	return renderError(lang, back, i18n.Func(lang)(i18n.CommonErrorSummaryTitle), detail, detail)
}

func NotAcceptableHTML(lang language.Tag, back weblinking.Link, supported mediatypes.Types) ([]byte, error) {
	T := i18n.Func(lang)
	title := T(i18n.HTTPErrorNotAcceptable)
	meta := T(i18n.HTTPErrorNotAcceptableDescription)
	msg := strings.Join([]string{meta, T(i18n.HTTPErrorNotAcceptableAvailable, supported)}, " ")
	return renderError(lang, back, title, meta, msg)
}
