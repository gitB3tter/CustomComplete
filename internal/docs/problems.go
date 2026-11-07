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
	problems               path = "/docs/problems"
	problemGeneric         path = "/docs/problems/generic"
	problemValidation      path = "/docs/problems/validation"
	problemExtensionErrors path = "/docs/problems/extensions/errors"
)

var (
	//go:embed problems/extensions/errors.tmpl.json
	problemsExtensionErrorsTmplJSON string
	//go:embed problems/validation.tmpl.json
	problemsValidationTmplJSON string
	//go:embed problems/generic.tmpl.json
	problemsGenericTmplJSON string
	problemsTemplatesJSON   = map[path][]string{
		problemValidation:      {problemsValidationTmplJSON},
		problemGeneric:         {problemsGenericTmplJSON},
		problemExtensionErrors: {problemsExtensionErrorsTmplJSON},
	}
	problemPaths = []path{
		problems,
		problemGeneric,
		problemValidation,
		problemExtensionErrors,
	}
)

func (h *docsHandler) initProblems(baseTemplates *ui.BaseTemplates) (err error) {
	escapes := make(map[path]map[language.Tag]weblinking.Link)
	links := make(map[path]map[language.Tag]weblinking.Links)
	items := make(map[language.Tag]weblinking.Links)
	for _, path := range problemPaths {
		escapes[path] = make(map[language.Tag]weblinking.Link)
		links[path] = make(map[language.Tag]weblinking.Links)
	}
	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)
		escapes[problems][langTag] = weblinking.New(weblinking.Up, docsPath()).WithTitle(T(i18n.DocsTitlePrefix))
		links[problems][langTag] = append(links[problems][langTag],
			escapes[problems][langTag],
			weblinking.New(weblinking.Self, problems.url()),
		)
		items[langTag] = weblinking.Links{
			weblinking.New(weblinking.Item, problemValidation.url()).WithTitle(T(i18n.NavigationDocsProblemValidation)).WithType(mediatypes.SchemaJSON),
			weblinking.New(weblinking.Item, problemGeneric.url()).WithTitle(T(i18n.NavigationDocsProblemGeneric)).WithType(mediatypes.SchemaJSON),
		}
		escapes[problemValidation][langTag] = weblinking.New(weblinking.Up, problems.url()).WithTitle(T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationDocsProblems))
		escapes[problemExtensionErrors][langTag] = weblinking.New(weblinking.Up, problems.url()).WithTitle(T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationDocsProblems))
		escapes[problemGeneric][langTag] = weblinking.New(weblinking.Up, problems.url()).WithTitle(T(i18n.DocsTitlePrefix) + ": " + T(i18n.NavigationDocsProblems))
	}

	for path := range problemsTemplatesJSON {
		h.cache[path], err = transport.New(path.url(), escapes[path], links[path]).With(mediatypes.SchemaJSON, baseTemplates, problemsTemplatesJSON[path], nil)
		if err != nil {
			return
		}
	}

	if h.cache[problems], err = transport.New(problems.url(), escapes[problems], links[problems]).WithLinkset(items, mediatypes.Linkset, mediatypes.LinksetJSON); err != nil {
		return
	}
	return
}
