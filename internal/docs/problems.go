package docs

import (
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/weblinking"
	_ "embed"
	"encoding/json"

	"golang.org/x/text/language"
)

var (
	problemsPaths []weblinking.Path = []weblinking.Path{
		transport.ProblemGenericPath,
		transport.ProblemErrorsPath,
		errs.SchemaPath,
	}
)

func (h *docsHandler) initProblems() (err error) {
	escapes := make(map[weblinking.Path]map[language.Tag]weblinking.Link)
	links := make(map[weblinking.Path]map[language.Tag]weblinking.Links)
	for _, path := range problemsPaths {
		escapes[path] = make(map[language.Tag]weblinking.Link)
		links[path] = make(map[language.Tag]weblinking.Links)
	}
	for _, langTag := range i18n.SupportedLanguages {
		T := i18n.Func(langTag)
		escapes[transport.ProblemErrorsPath][langTag] = weblinking.New(weblinking.Up, weblinking.DocsPath).
			WithTitle(T(i18n.NavigationDocumentation))
		links[transport.ProblemErrorsPath][langTag] = weblinking.Links{
			weblinking.New(weblinking.Up, weblinking.DocsPath).
				WithTitle(T(i18n.NavigationDocumentation)),
			weblinking.New(weblinking.Self, transport.ProblemErrorsPath),
		}
		escapes[transport.ProblemGenericPath][langTag] = weblinking.New(weblinking.Up, weblinking.DocsPath).
			WithTitle(T(i18n.NavigationDocumentation))
		links[transport.ProblemGenericPath][langTag] = weblinking.Links{
			weblinking.New(weblinking.Up, weblinking.DocsPath).
				WithTitle(T(i18n.NavigationDocumentation)),
			weblinking.New(weblinking.Self, transport.ProblemGenericPath),
		}
		escapes[errs.SchemaPath][langTag] = weblinking.New(weblinking.Up, transport.ProblemErrorsPath).
			WithTitle(T(i18n.NavigationDocumentation) + ": " + T(i18n.NavigationDocsProblemsError))
		links[errs.SchemaPath][langTag] = weblinking.Links{
			weblinking.New(weblinking.Up, transport.ProblemErrorsPath).
				WithTitle(T(i18n.NavigationDocumentation) + ": " + T(i18n.NavigationDocsProblemsError)),
			weblinking.New(weblinking.Self, errs.SchemaPath),
		}
	}

	output := map[weblinking.Path]map[language.Tag][]byte{
		transport.ProblemErrorsPath:  make(map[language.Tag][]byte, len(i18n.SupportedLanguages)),
		transport.ProblemGenericPath: make(map[language.Tag][]byte, len(i18n.SupportedLanguages)),
		errs.SchemaPath:              make(map[language.Tag][]byte, len(i18n.SupportedLanguages)),
	}
	for _, langTag := range i18n.SupportedLanguages {
		if output[transport.ProblemErrorsPath][langTag], err = json.Marshal(transport.ProblemsErrorsSchema(langTag)); err != nil {
			return
		}
		if output[transport.ProblemGenericPath][langTag], err = json.Marshal(transport.ProblemGenericSchema(langTag)); err != nil {
			return
		}
		if output[errs.SchemaPath][langTag], err = json.Marshal(errs.Schema(langTag)); err != nil {
			return
		}
	}

	for _, path := range problemsPaths {
		h.cache[path] = transport.NewCache(
			path,
			escapes[path],
			links[path],
		).With(
			mediatypes.SchemaJSON,
			output[path],
		)

	}

	return
}
