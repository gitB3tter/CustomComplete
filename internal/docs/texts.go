package docs

import (
	"customcomplete/internal/home"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"fmt"

	"golang.org/x/text/language"
)

const (
	TextsIndex      path = "/docs/texts"
	TextsItem       path = "/docs/texts/item"
	TextsCollection path = "/docs/texts/collection"
)

var (
	//go:embed texts/item.tmpl.html
	textItemTmplHTML string
	//go:embed texts/collection.tmpl.html
	textCollectionTmplHTML string

	templatesHTML = map[path][]string{
		TextsItem:       {textItemTmplHTML},
		TextsIndex:      {collectionHTML},
		TextsCollection: {textCollectionTmplHTML},
	}

	texts = []path{
		TextsItem,
		TextsIndex,
		TextsCollection,
	}
)

// has to be called after json to read out metadata
func (h *docsHandler) initTexts(baseTemplates *ui.BaseTemplates) (err error) {

	data := make(map[path]map[language.Tag]any)
	links := make(map[path]map[language.Tag]weblinking.Links)

	for _, path := range texts {
		data[path] = make(map[language.Tag]any)
		links[path] = make(map[language.Tag]weblinking.Links)
	}
	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)

		breadcrumbs := weblinking.Links{
			weblinking.Breadcrumb(home.Path(), T(i18n.NavigationHome)),
		}

		home := home.Path()

		links[TextsIndex][langTag] = weblinking.Links{
			weblinking.New(weblinking.Item, home.JoinPath(string(TextsCollection))).WithTitle(T(i18n.NavigationDocsTextsCollection)),
			weblinking.New(weblinking.Item, home.JoinPath(string(TextsItem))).WithTitle(T(i18n.NavigationDocsTextsItem)),
		}

		data[TextsIndex][langTag] = collection{
			Base: ui.BaseDataHTML{
				Title:       T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationTextCollection),
				Breadcrumbs: append(breadcrumbs, weblinking.New(weblinking.Up, docsPath()).WithTitle(T(i18n.NavigationDocumentation))),
				Lang:        langTag,
			},
			Items: links[TextsIndex][langTag],
		}

		links[TextsIndex][langTag] = append(append(links[TextsIndex][langTag], breadcrumbs...),
			weblinking.New(weblinking.Up, docsPath()).WithTitle(T(i18n.NavigationDocumentation)),
			weblinking.New(weblinking.Self, home.JoinPath(string(TextsIndex))).WithTitle(T(i18n.DocsTitlePrefix)+": "+T(i18n.NavigationDocsTextsCollection)),
		)

		breadcrumbs = append(breadcrumbs,
			weblinking.Breadcrumb(docsPath(), T(i18n.NavigationDocumentation)),
			weblinking.New(weblinking.Up, docsPath().JoinPath("texts")).WithTitle(T(i18n.NavigationTextCollection)),
		)

		links[TextsItem][langTag] = append(breadcrumbs, weblinking.New(weblinking.Self, docsPath().JoinPath("texts", "item")))

		filesizeSchema, ok := h.cache[SchemaTextsItem].GetSize(mediatypes.SchemaJSON, langTag)
		if !ok {
			return fmt.Errorf("filesize schema textitem not found")
		}
		filesizeSpec, ok := h.cache[SpecsTextsItem].GetSize(mediatypes.OpenAPIJSON, langTag)
		if !ok {
			return fmt.Errorf("filesize spec textitem not found")
		}

		data[TextsItem][langTag] = struct {
			Base   ui.BaseDataHTML
			Schema attachment
			Spec   attachment
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationDocsTextsItem),
				MetaDescription: T(i18n.DocsTitlePrefix) + ": " + T(i18n.TextItemMetaDescription),
				Lang:            langTag,
				Breadcrumbs:     breadcrumbs,
			},
			Schema: attachment{
				Path:        home.JoinPath(string(SchemaTextsItem)),
				Filename:    "Schema",
				FileSizeFmt: i18n.FormatBytesLocale(int64(filesizeSchema), langTag),
				Type:        mediatypes.SchemaJSON,
			},
			Spec: attachment{
				Path:        home.JoinPath(string(SpecsTextsItem)),
				Filename:    "Spec",
				FileSizeFmt: i18n.FormatBytesLocale(int64(filesizeSpec), langTag),
				Type:        mediatypes.OpenAPIJSON,
			},
		}

		links[TextsCollection][langTag] = append(breadcrumbs, weblinking.New(weblinking.Self, docsPath().JoinPath("texts", "collection")))

		filesizeSpec, ok = h.cache[SpecsTextsCollection].GetSize(mediatypes.OpenAPIJSON, langTag)
		if !ok {
			return fmt.Errorf("filesize schema textcollection not found")
		}

		data[TextsCollection][langTag] = struct {
			Base   ui.BaseDataHTML
			Schema attachment
			Spec   attachment
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationDocsTextsCollection),
				MetaDescription: T(i18n.DocsTitlePrefix) + ": " + T(i18n.TextCollectionMetaDescription),
				Lang:            langTag,
				Breadcrumbs:     breadcrumbs,
			},
			Spec: attachment{
				Path:        home.JoinPath(string(SpecsTextsCollection)),
				Filename:    "Spec",
				FileSizeFmt: i18n.FormatBytesLocale(int64(filesizeSpec), langTag),
				Type:        mediatypes.OpenAPIJSON,
			},
		}

	}
	for path, date := range data {
		h.cache[path], err = transport.New(
			path.url(),
			map[language.Tag]weblinking.Link{
				language.English: weblinking.New(weblinking.Up, home.Path()),
				language.German:  weblinking.New(weblinking.Up, home.Path()),
			},
			links[path],
		).With(mediatypes.HTML, baseTemplates, templatesHTML[path], date)
		if err != nil {
			return
		}
	}

	h.cache[TextsIndex], err = h.cache[TextsIndex].WithLinkset(links[TextsIndex],
		supportedTypesCollection...,
	)

	return
}
