package docs

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/home"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	SpecsMain                         path = "/docs/specs"
	SpecsTextsItem                    path = "/docs/specs/texts/item"
	SpecsTextsCollection              path = "/docs/specs/texts/collection"
	SpecsTextsFormBody                path = "/docs/specs/texts/components/form-body"
	SpecsHeadersEtagStrong            path = "/docs/specs/components/headers/etag-strong"
	SpecsHeadersEtagWeak              path = "/docs/specs/components/headers/etag-weak"
	SpecsHeadersNoCache               path = "/docs/specs/components/headers/no-cache"
	SpecsHeadersNoStore               path = "/docs/specs/components/headers/no-store"
	SpecsHeadersReturnApplied         path = "/docs/specs/components/headers/preference-return-applied"
	SpecsParamsLanguage               path = "/docs/specs/components/parameters/accept-language"
	SpecsParamsPreferReturn           path = "/docs/specs/components/parameters/prefer-return"
	SpecsParamsIfMatch                path = "/docs/specs/components/parameters/if-match"
	SpecsResponsesNoBody              path = "/docs/specs/components/responses/body-not-allowed"
	SpecsResponsesNotParseable        path = "/docs/specs/components/responses/body-not-parseable"
	SpecsResponsesTooLarge            path = "/docs/specs/components/responses/content-too-large"
	SpecsResponsesInternalError       path = "/docs/specs/components/responses/internal-error"
	SpecsResponsesNotAcceptable       path = "/docs/specs/components/responses/not-acceptable"
	SpecsResponsesNotFound            path = "/docs/specs/components/responses/not-found"
	SpecsResponsePreconditionFailed   path = "/docs/specs/components/responses/precondition-failed"
	SpecsResponsePreconditionRequired path = "/docs/specs/components/responses/precondition-required"
	SpecsResponseUnsupportedMediaType path = "/docs/specs/components/responses/unsupported-mediatype"
	SpecsResponsesUnprocessable       path = "/docs/specs/components/responses/unprocessable-content"
)

var (
	//go:embed specs/main.tmpl.json
	specMainTmplJSON string

	//go:embed specs/texts/item.tmpl.json
	specTextItemTmplJSON string
	//go:embed specs/texts/collection.tmpl.json
	specTextCollectionTmplJSON string

	//go:embed specs/texts/components/form-body.tmpl.json
	specsTextsFormBody string

	//go:embed specs/components/headers/etag-strong.tmpl.json
	specsHeadersEtagStrong string
	//go:embed specs/components/headers/etag-weak.tmpl.json
	specsHeadersEtagWeak string
	//go:embed specs/components/headers/no-cache.tmpl.json
	specsHeadersNoCache string
	//go:embed specs/components/headers/no-store.tmpl.json
	specsHeadersNoStore string
	//go:embed specs/components/headers/preference-return-applied.tmpl.json
	specsHeadersReturnApplied string

	//go:embed specs/components/parameters/accept-language.tmpl.json
	specsParamsLanguage string
	//go:embed specs/components/parameters/prefer-return.tmpl.json
	specsParamsPreferReturn string
	//go:embed specs/components/parameters/if-match.tmpl.json
	specsParamsIfMatch string

	//go:embed specs/components/responses/body-not-allowed.tmpl.json
	specsResponsesNoBody string
	//go:embed specs/components/responses/body-not-parseable.tmpl.json
	specsResponsesNotParseable string
	//go:embed specs/components/responses/content-too-large.tmpl.json
	specsResponsesTooLarge string
	//go:embed specs/components/responses/internal-error.tmpl.json
	specsResponsesInternalError string
	//go:embed specs/components/responses/not-acceptable.tmpl.json
	specsResponsesNotAcceptable string
	//go:embed specs/components/responses/not-found.tmpl.json
	specsResponsesNotFound string
	//go:embed specs/components/responses/precondition-failed.tmpl.json
	specsResponsePreconditionFailed string
	//go:embed specs/components/responses/precondition-required.tmpl.json
	specsResponsePreconditionRequired string
	//go:embed specs/components/responses/unsupported-mediatype.tmpl.json
	specsResponseUnsupportedMediaType string
	//go:embed specs/components/responses/unprocessable-content.tmpl.json
	specsResponsesUnprocessable string

	specTemplatesSchemasSchemaJSON = map[path][]string{
		SpecsMain:                         {specMainTmplJSON},
		SpecsTextsItem:                    {specTextItemTmplJSON},
		SpecsTextsCollection:              {specTextCollectionTmplJSON},
		SpecsTextsFormBody:                {specsTextsFormBody},
		SpecsHeadersEtagStrong:            {specsHeadersEtagStrong},
		SpecsHeadersEtagWeak:              {specsHeadersEtagWeak},
		SpecsHeadersNoCache:               {specsHeadersNoCache},
		SpecsHeadersNoStore:               {specsHeadersNoStore},
		SpecsHeadersReturnApplied:         {specsHeadersReturnApplied},
		SpecsParamsLanguage:               {specsParamsLanguage},
		SpecsParamsPreferReturn:           {specsParamsPreferReturn},
		SpecsResponsesNoBody:              {specsResponsesNoBody},
		SpecsParamsIfMatch:                {specsParamsIfMatch},
		SpecsResponsesNotParseable:        {specsResponsesNotParseable},
		SpecsResponsesTooLarge:            {specsResponsesTooLarge},
		SpecsResponsesInternalError:       {specsResponsesInternalError},
		SpecsResponsesNotAcceptable:       {specsResponsesNotAcceptable},
		SpecsResponsesNotFound:            {specsResponsesNotFound},
		SpecsResponsePreconditionFailed:   {specsResponsePreconditionFailed},
		SpecsResponsePreconditionRequired: {specsResponsePreconditionRequired},
		SpecsResponseUnsupportedMediaType: {specsResponseUnsupportedMediaType},
		SpecsResponsesUnprocessable:       {specsResponsesUnprocessable},
	}
)

func (h *docsHandler) initSpecs(baseTemplates *ui.BaseTemplates) (err error) {
	escapes := make(map[path]map[language.Tag]weblinking.Link)
	links := make(map[path]map[language.Tag]weblinking.Links)
	data := make(map[path]map[language.Tag]any)
	for path := range specTemplatesSchemasSchemaJSON {
		escapes[path] = make(map[language.Tag]weblinking.Link)
		links[path] = make(map[language.Tag]weblinking.Links)
		data[path] = make(map[language.Tag]any)
	}
	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)
		for path := range specTemplatesSchemasSchemaJSON {
			escapes[path][langTag] = weblinking.New(weblinking.Up, home.Path().JoinPath(string(docs))).WithTitle(T(i18n.NavigationDocumentation))
			data[path][langTag] = struct {
				WordMaxFmt      string
				FilesizeLimitMB int64
			}{
				WordMaxFmt:      message.NewPrinter(langTag).Sprintf("%d", 100000),
				FilesizeLimitMB: catalog.MaxFileSizeMB,
			}
		}
	}

	for path := range escapes {
		h.cache[path], err = transport.New(path.url(), escapes[path], links[path]).With(mediatypes.OpenAPIJSON, baseTemplates, specTemplatesSchemasSchemaJSON[path], data[path])
		if err != nil {
			return
		}
	}
	return
}
