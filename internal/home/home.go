package home

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
	//go:embed home.html
	tmpl string
)

type homehandler struct {
	cache transport.PrebakedCache
}

func (h *homehandler) Register(mux *http.ServeMux) {
	h.cache.RegisterOn(mux, "/{$}")
}

func NewHandler(baseTemplates *ui.BaseTemplates) (*homehandler, error) {

	handler := &homehandler{}

	home := weblinking.RootPath

	links := make(map[language.Tag]weblinking.Links)
	escape := make(map[language.Tag]weblinking.Link)
	data := make(map[language.Tag]any)
	for _, lang := range i18n.SupportedLanguages {
		T := i18n.Func(lang)

		texts := weblinking.New(weblinking.Item, home.Join("texts")).WithTitle(T(i18n.NavigationTextCollection))
		suggestions := weblinking.New(weblinking.Item, home.Join("suggestions")).WithTitle(T(i18n.NavigationSuggestions))
		docs := weblinking.New(weblinking.Item, home.Join("docs")).WithTitle(T(i18n.NavigationDocumentation))
		serviceDoc := weblinking.New(weblinking.ServiceDoc, home.Join("docs")).WithTitle(T(i18n.NavigationDocumentation))
		serviceDesc := weblinking.New(weblinking.ServiceDesc, home.Join("docs", "specs")).WithTitle(T(i18n.NavigationDocumentation))

		escape[lang] = weblinking.New(weblinking.Up, home)

		links[lang] = weblinking.Links{
			texts.WithType(mediatypes.HTML),
			texts.WithType(mediatypes.Linkset),
			texts.WithType(mediatypes.LinksetJSON),
			suggestions.WithType(mediatypes.HTML),
			suggestions.WithType(mediatypes.Linkset),
			suggestions.WithType(mediatypes.LinksetJSON),
			docs.WithType(mediatypes.HTML),
			serviceDoc.WithType(mediatypes.HTML),
			serviceDesc.WithType(mediatypes.OpenAPIJSON),
		}

		data[lang] = struct {
			Base ui.BaseDataHTML
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationHome),
				Lang:            lang,
				MetaDescription: T(i18n.HomeMetaDescription),
				Links:           links[lang],
			},
		}
	}

	var err error
	var content ui.Prebaked

	if content, err = baseTemplates.Prebake(mediatypes.HTML, []string{tmpl}, data); err != nil {
		return nil, err
	} else {
		handler.cache = transport.NewCache(weblinking.RootPath, escape, links).With(mediatypes.HTML, content)
	}

	if handler.cache, err = handler.cache.WithLinkset(links); err != nil {
		return nil, err
	}

	return handler, nil
}
