package ui

import (
	"bytes"
	"customcomplete/internal/i18n"
	"customcomplete/internal/weblinking"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/url"

	"golang.org/x/text/language"
)

type BaseDataHTML struct {
	Title           string
	MetaDescription string
	Lang            language.Tag
	NoIndex         bool
	HasError        bool
	Breadcrumbs     weblinking.Links
	Back            weblinking.Link
	Links           weblinking.Links
}

const (
	nameBase        = "base.html"
	HiddenFieldEtag = "If-Match"
)

var (
	//go:embed html.html
	templateBaseHTML string
	//go:embed errors.html
	templateErrorsHTML string
	//go:embed redirect.html
	templateRedirectHTML string
	//go:embed html_partials.html
	templatePartialsHTML string
)

func buildBaseTemplatesHTML() (map[language.Tag]*template.Template, error) {

	baseTemplates := make(map[language.Tag]*template.Template)
	for _, lang := range i18n.SupportedLanguages {
		basePathRel := (&url.URL{Path: "/"}).JoinPath("docs", "rels")
		funcs := template.FuncMap{
			"T": i18n.TemplateTranslate(lang),
			"relURL": func(rel string) string {
				return basePathRel.JoinPath(rel).String()
			},
		}

		template, err := template.New(nameBase).Funcs(funcs).Parse(templateBaseHTML)
		if err != nil {
			return nil, err
		}
		if _, err = template.New("partials").Parse(templatePartialsHTML); err != nil {
			return nil, err
		}

		baseTemplates[lang] = template
	}
	return baseTemplates, nil
}

func (b *BaseTemplates) WithHTML(tmpls ...string) error {
	for lang, template := range b.html {
		for _, tmplString := range tmpls {
			if _, err := template.New("").Parse(tmplString); err != nil {
				return err
			}
		}
		b.html[lang] = template
	}
	return nil
}

func (baseTemplates *BaseTemplates) BuildHTML(templates ...string) (map[language.Tag]*template.Template, error) {

	tmpls := make(map[language.Tag]*template.Template)

	for lang, baseTemplate := range baseTemplates.html {
		clone, err := baseTemplate.Clone()
		if err != nil {
			return nil, err
		}

		for _, template := range templates {
			clone, err = clone.New("").Parse(template)
			if err != nil {
				return nil, err
			}
		}

		tmpls[lang] = clone
	}

	return tmpls, nil

}

func BufferHTML(tmpl *template.Template, data any) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, nameBase, data); err != nil {
		log.Printf("template error: %v", err)
		return nil, fmt.Errorf(`cant buffer html template "%s": %w`, nameBase, err)
	}
	return buf.Bytes(), nil
}

func (b *BaseTemplates) prebakeHTML(templates []string, data map[language.Tag]any) (map[language.Tag][]byte, error) {
	tmpls, err := b.BuildHTML(templates...)
	if err != nil {
		return nil, err
	}
	out := make(map[language.Tag][]byte, len(tmpls))
	for lang, tmpl := range tmpls {
		buf, err := BufferHTML(tmpl, data[lang])
		if err != nil {
			return nil, err
		}
		out[lang] = buf
	}
	return out, nil
}
