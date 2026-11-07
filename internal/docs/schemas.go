package docs

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"net/url"

	"golang.org/x/text/language"
)

const (
	SchemaTextsItem       path = "/docs/schemas/texts/item"
	SchemaTextsItemStatus path = "/docs/schemas/texts/item/status"
)

var (
	//go:embed schemas/texts/item.tmpl.json
	schemaTextItemTmplJSON string
	//go:embed schemas/texts/item/status.tmpl.json
	schemaTextItemStatusTmplJSON string
	schemaTemplatesSchemaJSON    = map[path][]string{
		SchemaTextsItem:       {schemaTextItemTmplJSON},
		SchemaTextsItemStatus: {schemaTextItemStatusTmplJSON},
	}
)

func (h *docsHandler) initSchemas(baseTemplates *ui.BaseTemplates) (err error) {
	escapes := make(map[path]map[language.Tag]weblinking.Link)
	links := make(map[path]map[language.Tag]weblinking.Links)
	docsPath := (&url.URL{Path: "/"}).JoinPath("docs")
	for path := range schemaTemplatesSchemaJSON {
		escapes[path] = make(map[language.Tag]weblinking.Link)
		links[path] = make(map[language.Tag]weblinking.Links)
		for _, langTag := range i18n.SupportedLanguages {
			T := i18n.Func(langTag)
			escapes[path][langTag] = weblinking.New(weblinking.Up, docsPath.JoinPath("texts", "item")).WithTitle(T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationTextCollection))
		}
		h.cache[path] = transport.New(path.url(), escapes[path], links[path])
	}

	dataSingle := struct {
		NameMaxCodePoints int
	}{
		NameMaxCodePoints: catalog.DocumentNameMaxCodePoints,
	}
	data := make(map[language.Tag]any)
	for _, langTag := range i18n.SupportedLanguages {
		data[langTag] = dataSingle
	}

	if h.cache[SchemaTextsItem], err = h.cache[SchemaTextsItem].With(mediatypes.SchemaJSON, baseTemplates, schemaTemplatesSchemaJSON[SchemaTextsItem], data); err != nil {
		return
	}
	h.cache[SchemaTextsItemStatus], err = h.cache[SchemaTextsItemStatus].With(mediatypes.SchemaJSON, baseTemplates, schemaTemplatesSchemaJSON[SchemaTextsItemStatus], nil)
	return
}
