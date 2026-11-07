package docs

import (
	"customcomplete/internal/home"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"net/http"
	"net/url"

	"golang.org/x/text/language"
)

var (
	//go:embed collection.html
	collectionHTML string
	//go:embed docs.html
	docsHTML                 string
	supportedTypesCollection mediatypes.Types = mediatypes.Types{mediatypes.HTML, mediatypes.LinksetJSON, mediatypes.Linkset}
)

type path string

func (p path) url() *url.URL {
	return &url.URL{Path: string(p)}
}

const (
	docs path = "/docs"
)

type docsHandler struct {
	cache map[path]transport.PrebakedCache
}

func (h *docsHandler) Register(mux *http.ServeMux) {
	for _, prebaked := range h.cache {
		prebaked.RegisterOn(mux, "")
	}
}

func NewHandler(baseTemplates *ui.BaseTemplates) (*docsHandler, error) {

	handler := &docsHandler{
		cache: make(map[path]transport.PrebakedCache),
	}

	if err := baseTemplates.WithHTML(docsHTML); err != nil {
		return nil, err
	}
	if err := handler.initSchemas(baseTemplates); err != nil {
		return nil, err
	}
	if err := handler.initSpecs(baseTemplates); err != nil {
		return nil, err
	}
	if err := handler.initLinkRels(baseTemplates); err != nil {
		return nil, err
	}
	if err := handler.initProblems(baseTemplates); err != nil {
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
		escapes[langTag] = weblinking.New(weblinking.Up, home.Path()).WithTitle(T(i18n.NavigationHome))

		home := home.Path()

		links[langTag] = weblinking.Links{
			weblinking.New(weblinking.Item, home.JoinPath(string(TextsIndex))).WithTitle(T(i18n.NavigationTextCollection)),
			weblinking.New(weblinking.Item, home.JoinPath(string(LinkRels))).WithTitle(T(i18n.NavigationDocsRels)),
		}

		homeLink := weblinking.New(weblinking.Up, home).WithTitle(T(i18n.NavigationHome))

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
			weblinking.New(weblinking.Self, home.JoinPath(string(docs))).WithTitle(T(i18n.NavigationDocumentation)),
			homeLink,
		)
	}
	h.cache[docs], err = transport.New(docs.url(), escapes, links).With(mediatypes.HTML, baseTemplates, []string{collectionHTML}, data)
	if err != nil {
		return
	}

	h.cache[docs], err = h.cache[docs].WithLinkset(links,
		supportedTypesCollection...,
	)

	return

}
