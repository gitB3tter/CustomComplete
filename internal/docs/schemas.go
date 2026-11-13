package docs

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/texts"
	"customcomplete/internal/transport"
	"customcomplete/internal/weblinking"
	"encoding/json"

	"golang.org/x/text/language"
)

func (h *docsHandler) initSchemas() (err error) {
	escapes := make(map[weblinking.Path]map[language.Tag]weblinking.Link)
	links := make(map[weblinking.Path]map[language.Tag]weblinking.Links)
	for _, path := range []weblinking.Path{texts.ItemSchemaPath, texts.StatusSchemaPath} {
		escapes[path] = make(map[language.Tag]weblinking.Link)
		links[path] = make(map[language.Tag]weblinking.Links)
		for _, langTag := range i18n.SupportedLanguages {
			T := i18n.Func(langTag)
			escapes[path][langTag] = weblinking.New(weblinking.Up, weblinking.DocsPath).WithTitle(T(i18n.NavigationDocumentation))
		}
		h.cache[path] = transport.NewCache(path, escapes[path], links[path])
	}

	content := map[weblinking.Path]map[language.Tag][]byte{
		texts.ItemSchemaPath:   make(map[language.Tag][]byte, len(i18n.SupportedLanguages)),
		texts.StatusSchemaPath: make(map[language.Tag][]byte, len(i18n.SupportedLanguages)),
	}
	for _, langTag := range i18n.SupportedLanguages {
		if content[texts.ItemSchemaPath][langTag], err = json.Marshal(texts.ItemSchema(langTag)); err != nil {
			return
		}
		if content[texts.StatusSchemaPath][langTag], err = json.Marshal(texts.StatusSchema(langTag)); err != nil {
			return
		}
	}

	h.cache[texts.ItemSchemaPath] = h.cache[texts.ItemSchemaPath].
		With(
			mediatypes.SchemaJSON,
			content[texts.ItemSchemaPath],
		)
	h.cache[texts.StatusSchemaPath] = h.cache[texts.StatusSchemaPath].
		With(
			mediatypes.SchemaJSON,
			content[texts.StatusSchemaPath],
		)
	return
}
