package docs

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/texts"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"fmt"

	"golang.org/x/text/language"
)

const (
	TextsIndexPath weblinking.Path = "/docs/texts"
)

var (
	//go:embed texts/item.tmpl.html
	textItemTmplHTML string
	//go:embed texts/collection.tmpl.html
	textCollectionTmplHTML string

	templatesHTML = map[weblinking.Path][]string{
		texts.ItemDocPath:       {textItemTmplHTML},
		TextsIndexPath:          {collectionHTML},
		texts.CollectionDocPath: {textCollectionTmplHTML},
	}
)

// has to be called after json to read out metadata
func (h *docsHandler) initTexts(baseTemplates *ui.BaseTemplates) (err error) {

	data := make(map[weblinking.Path]map[language.Tag]any)
	links := make(map[weblinking.Path]map[language.Tag]weblinking.Links)

	for path := range templatesHTML {
		data[path] = make(map[language.Tag]any)
		links[path] = make(map[language.Tag]weblinking.Links)
	}
	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)

		breadcrumbs := weblinking.Links{
			weblinking.Breadcrumb(weblinking.RootPath, T(i18n.NavigationHome)),
		}

		links[TextsIndexPath][langTag] = weblinking.Links{
			weblinking.New(weblinking.Item, texts.CollectionDocPath).WithTitle(T(i18n.NavigationDocsTextsCollection)),
			weblinking.New(weblinking.Item, texts.ItemDocPath).WithTitle(T(i18n.NavigationDocsTextsItem)),
		}

		data[TextsIndexPath][langTag] = collection{
			Base: ui.BaseDataHTML{
				Title:       T(i18n.NavigationDocumentation) + ": " + T(i18n.NavigationTextCollection),
				Breadcrumbs: append(breadcrumbs, weblinking.New(weblinking.Up, weblinking.DocsPath).WithTitle(T(i18n.NavigationDocumentation))),
				Lang:        langTag,
			},
			Items: links[TextsIndexPath][langTag],
		}

		links[TextsIndexPath][langTag] = append(append(links[TextsIndexPath][langTag], breadcrumbs...),
			weblinking.New(weblinking.Up, weblinking.DocsPath).WithTitle(T(i18n.NavigationDocumentation)),
			weblinking.New(weblinking.Self, TextsIndexPath).WithTitle(T(i18n.NavigationDocumentation)+": "+T(i18n.NavigationDocsTextsCollection)),
		)

		breadcrumbs = append(breadcrumbs,
			weblinking.New(weblinking.Up, weblinking.DocsPath).WithTitle(T(i18n.NavigationDocumentation)),
			weblinking.New(weblinking.Up, TextsIndexPath).WithTitle(T(i18n.NavigationTextCollection)),
		)

		links[texts.ItemDocPath][langTag] = append(breadcrumbs, weblinking.New(weblinking.Self, texts.ItemDocPath))

		filesizeSchema, ok := h.cache[texts.ItemSchemaPath].GetSize(mediatypes.SchemaJSON, langTag)
		if !ok {
			return fmt.Errorf("filesize schema textitem not found")
		}
		filesizeSpec, ok := h.cache[texts.ItemSpecPath].GetSize(mediatypes.OpenAPIJSON, langTag)
		if !ok {
			return fmt.Errorf("filesize spec textitem not found")
		}

		data[texts.ItemDocPath][langTag] = struct {
			Base   ui.BaseDataHTML
			Schema attachment
			Spec   attachment
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationDocumentation) + ": " + T(i18n.NavigationDocsTextsItem),
				MetaDescription: T(i18n.NavigationDocumentation) + ": " + T(i18n.TextItemMetaDescription),
				Lang:            langTag,
				Breadcrumbs:     breadcrumbs,
			},
			Schema: attachment{
				Path:        texts.ItemSchemaPath,
				Filename:    "Schema",
				FileSizeFmt: i18n.FormatBytesLocale(int64(filesizeSchema), langTag),
				Type:        mediatypes.SchemaJSON,
			},
			Spec: attachment{
				Path:        texts.ItemSpecPath,
				Filename:    "Spec",
				FileSizeFmt: i18n.FormatBytesLocale(int64(filesizeSpec), langTag),
				Type:        mediatypes.OpenAPIJSON,
			},
		}

		links[texts.CollectionDocPath][langTag] = append(breadcrumbs, weblinking.New(weblinking.Self, texts.CollectionDocPath))

		filesizeSpec, ok = h.cache[texts.CollectionSpecPath].GetSize(mediatypes.OpenAPIJSON, langTag)
		if !ok {
			return fmt.Errorf("filesize schema textcollection not found")
		}

		data[texts.CollectionDocPath][langTag] = struct {
			Base   ui.BaseDataHTML
			Schema attachment
			Spec   attachment
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationDocumentation) + ": " + T(i18n.NavigationDocsTextsCollection),
				MetaDescription: T(i18n.NavigationDocumentation) + ": " + T(i18n.TextCollectionMetaDescription),
				Lang:            langTag,
				Breadcrumbs:     breadcrumbs,
			},
			Spec: attachment{
				Path:        texts.CollectionSpecPath,
				Filename:    "Spec",
				FileSizeFmt: i18n.FormatBytesLocale(int64(filesizeSpec), langTag),
				Type:        mediatypes.OpenAPIJSON,
			},
		}

	}
	for path, date := range data {
		if content, err := baseTemplates.Prebake(mediatypes.HTML, templatesHTML[path], date); err != nil {
			return err
		} else {
			h.cache[path] = transport.NewCache(
				path,
				map[language.Tag]weblinking.Link{
					language.English: weblinking.New(weblinking.Up, weblinking.DocsPath),
					language.German:  weblinking.New(weblinking.Up, weblinking.DocsPath),
				},
				links[path],
			).With(mediatypes.HTML, content)
		}
	}

	h.cache[TextsIndexPath], err = h.cache[TextsIndexPath].WithLinkset(links[TextsIndexPath])

	return
}
