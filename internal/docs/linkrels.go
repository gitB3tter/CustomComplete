package docs

import (
	"customcomplete/internal/home"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"

	"golang.org/x/text/language"
)

const (
	LinkRels          path = "/docs/rels"
	LinkRelDeleteForm path = "/docs/rels/delete-form"
)

var (
	//go:embed linkrels/delete-form.tmpl.html
	linkRelsDeleteFormHTML string
	linkRelsTemplatesHTML  = map[path][]string{
		LinkRelDeleteForm: {linkRelsDeleteFormHTML},
		LinkRels:          {collectionHTML},
	}
)

func (h *docsHandler) initLinkRels(baseTemplates *ui.BaseTemplates) (err error) {
	escapes := make(map[path]map[language.Tag]weblinking.Link)
	links := make(map[path]map[language.Tag]weblinking.Links)
	data := make(map[path]map[language.Tag]any)
	for path := range linkRelsTemplatesHTML {
		escapes[path] = make(map[language.Tag]weblinking.Link)
		links[path] = make(map[language.Tag]weblinking.Links)
		data[path] = make(map[language.Tag]any)
	}

	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)
		escapes[LinkRelDeleteForm][langTag] = weblinking.New(weblinking.Up, docsPath().JoinPath("texts", "item")).WithTitle(T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationTextCollection))
		data[LinkRelDeleteForm][langTag] = struct {
			Base ui.BaseDataHTML
		}{
			Base: ui.BaseDataHTML{
				Title: T(i18n.NavigationDocsDeleteForm),
				Lang:  langTag,
				Breadcrumbs: weblinking.Links{
					weblinking.Breadcrumb(home.Path(), T(i18n.NavigationHome)),
					weblinking.Breadcrumb(docsPath(), T(i18n.NavigationDocumentation)),
					weblinking.New(weblinking.Up, LinkRels.url()).WithTitle(T(i18n.NavigationDocsRels)),
				},
			},
		}
		data[LinkRels][langTag] = collection{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationDocsRels),
				MetaDescription: T(i18n.DocsRelsCollectionMetaDescription),
				Lang:            langTag,
				Breadcrumbs: weblinking.Links{
					weblinking.Breadcrumb(home.Path(), T(i18n.NavigationHome)),
					weblinking.New(weblinking.Up, docsPath()).WithTitle(T(i18n.NavigationDocumentation)),
				},
			},
			Items: weblinking.Links{
				weblinking.New(weblinking.Item, GetPath(LinkRelDeleteForm)).WithTitle(T(i18n.NavigationDocsDeleteForm)),
			},
		}
	}

	for path := range escapes {
		h.cache[path] = transport.New(path.url(), escapes[path], links[path])
	}

	h.cache[LinkRels], err = h.cache[LinkRels].With(mediatypes.HTML, baseTemplates, linkRelsTemplatesHTML[LinkRels], data[LinkRels])
	if err != nil {
		return
	}
	h.cache[LinkRelDeleteForm], err = h.cache[LinkRelDeleteForm].With(mediatypes.HTML, baseTemplates, linkRelsTemplatesHTML[LinkRelDeleteForm], data[LinkRelDeleteForm])
	if err != nil {
		return
	}
	return
}
