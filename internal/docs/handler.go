package docs

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"net/http"

	"golang.org/x/text/language"
)

var (
	//go:embed collection.html
	collectionHTML string
	//go:embed docs.html
	docsHTML string
)

type docsHandler struct {
	cache map[weblinking.Path]transport.PrebakedCache
}

func (h *docsHandler) Register(mux *http.ServeMux) {
	for _, prebaked := range h.cache {
		prebaked.RegisterOn(mux, "")
	}
}

func NewHandler(baseTemplates *ui.BaseTemplates) (*docsHandler, error) {

	handler := &docsHandler{
		cache: make(map[weblinking.Path]transport.PrebakedCache),
	}

	if err := baseTemplates.WithHTML(docsHTML); err != nil {
		return nil, err
	}
	if err := handler.initSchemas(); err != nil {
		return nil, err
	}
	if err := handler.initSpecs(baseTemplates); err != nil {
		return nil, err
	}
	if err := handler.initLinkRels(baseTemplates); err != nil {
		return nil, err
	}
	if err := handler.initProblems(); err != nil {
		return nil, err
	}
	if err := handler.initTexts(baseTemplates); err != nil {
		return nil, err
	}
	if err := handler.initRoot(baseTemplates); err != nil {
		return nil, err
	}

	return handler, nil
}

func (h *docsHandler) initRoot(baseTemplates *ui.BaseTemplates) (err error) {

	escapes := make(map[language.Tag]weblinking.Link)

	data := make(map[language.Tag]any)
	links := make(map[language.Tag]weblinking.Links)

	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)

		links[langTag] = weblinking.Links{
			weblinking.New(weblinking.Item, TextsIndexPath).WithTitle(T(i18n.NavigationTextCollection)),
			weblinking.New(weblinking.Item, LinkRelsPath).WithTitle(T(i18n.NavigationDocsRels)),
		}

		homeLink := weblinking.New(weblinking.Up, weblinking.RootPath).WithTitle(T(i18n.NavigationHome))
		escapes[langTag] = homeLink

		data[langTag] = collection{
			Base: ui.BaseDataHTML{
				Title: T(i18n.NavigationDocumentation),
				Breadcrumbs: weblinking.Links{
					homeLink,
				},
				Lang: langTag,
			},
			Items: links[langTag],
		}

		links[langTag] = append(links[langTag],
			weblinking.New(weblinking.Self, weblinking.DocsPath).WithTitle(T(i18n.NavigationDocumentation)),
			homeLink,
		)
	}
	if contentDocs, err := baseTemplates.Prebake(mediatypes.HTML, []string{collectionHTML}, data); err != nil {
		return err
	} else {
		h.cache[weblinking.DocsPath] = transport.NewCache(weblinking.DocsPath, escapes, links).
			With(mediatypes.HTML, contentDocs)
	}

	h.cache[weblinking.DocsPath], err = h.cache[weblinking.DocsPath].WithLinkset(links)

	return

}
