package docs

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"

	"golang.org/x/text/language"
)

const (
	LinkRelsPath      weblinking.Path = "/docs/rels"
	LinkRelDeleteForm weblinking.Path = "/docs/rels/delete-form"
)

var (
	//go:embed linkrels/delete-form.tmpl.html
	linkRelsDeleteFormHTML string
	linkRelsTemplatesHTML  = map[weblinking.Path][]string{
		LinkRelDeleteForm: {linkRelsDeleteFormHTML},
		LinkRelsPath:      {collectionHTML},
	}
)

func (h *docsHandler) initLinkRels(baseTemplates *ui.BaseTemplates) (err error) {
	escapes := make(map[weblinking.Path]map[language.Tag]weblinking.Link)
	links := make(map[weblinking.Path]map[language.Tag]weblinking.Links)
	data := make(map[weblinking.Path]map[language.Tag]any)
	for path := range linkRelsTemplatesHTML {
		escapes[path] = make(map[language.Tag]weblinking.Link)
		links[path] = make(map[language.Tag]weblinking.Links)
		data[path] = make(map[language.Tag]any)
	}

	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)
		escapes[LinkRelDeleteForm][langTag] = weblinking.DocsPath.Link(weblinking.Up).WithTitle(T(i18n.NavigationDocumentation))
		data[LinkRelDeleteForm][langTag] = struct {
			Base ui.BaseDataHTML
		}{
			Base: ui.BaseDataHTML{
				Title: T(i18n.NavigationDocsDeleteForm),
				Lang:  langTag,
				Breadcrumbs: weblinking.Links{
					weblinking.Breadcrumb(weblinking.RootPath, T(i18n.NavigationHome)),
					weblinking.Breadcrumb(weblinking.DocsPath, T(i18n.NavigationDocumentation)),
					weblinking.New(weblinking.Up, LinkRelsPath).WithTitle(T(i18n.NavigationDocsRels)),
				},
			},
		}
		data[LinkRelsPath][langTag] = collection{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationDocsRels),
				MetaDescription: T(i18n.DocsRelsCollectionMetaDescription),
				Lang:            langTag,
				Breadcrumbs: weblinking.Links{
					weblinking.Breadcrumb(weblinking.RootPath, T(i18n.NavigationHome)),
					weblinking.New(weblinking.Up, weblinking.DocsPath).WithTitle(T(i18n.NavigationDocumentation)),
				},
			},
			Items: weblinking.Links{
				weblinking.New(weblinking.Item, LinkRelDeleteForm).WithTitle(T(i18n.NavigationDocsDeleteForm)),
			},
		}
	}

	if contentLinkRels, err := baseTemplates.Prebake(mediatypes.HTML, linkRelsTemplatesHTML[LinkRelsPath], data[LinkRelsPath]); err != nil {
		return err
	} else {
		h.cache[LinkRelsPath] = transport.NewCache(LinkRelsPath, escapes[LinkRelsPath], links[LinkRelsPath]).
			With(mediatypes.HTML, contentLinkRels)
	}

	if contentLinkRelDeleteForm, err := baseTemplates.Prebake(mediatypes.HTML, linkRelsTemplatesHTML[LinkRelDeleteForm], data[LinkRelDeleteForm]); err != nil {
		return err
	} else {
		h.cache[LinkRelDeleteForm] = transport.NewCache(LinkRelDeleteForm, escapes[LinkRelDeleteForm], links[LinkRelDeleteForm]).
			With(mediatypes.HTML, contentLinkRelDeleteForm)
	}

	return
}
