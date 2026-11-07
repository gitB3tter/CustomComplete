package ui

import (
	"fmt"
	html "html/template"
	text "text/template"

	"golang.org/x/text/language"
)

func init() {
	baseTemplates, err := NewBaseTemplates()
	if err != nil {
		panic(fmt.Errorf("build base template: %w", err))
	}
	templateErr, err = baseTemplates.BuildHTML(templateErrorsHTML)
	if err != nil {
		panic(fmt.Errorf("build error template: %w", err))
	}
	templateRedirect, err = baseTemplates.BuildHTML(templateRedirectHTML)
	if err != nil {
		panic(fmt.Errorf("build error template: %w", err))
	}
}

type BaseTemplates struct {
	html        map[language.Tag]*html.Template
	schemaJSON  map[language.Tag]*text.Template
	openApiJSON map[language.Tag]*text.Template
}

func NewBaseTemplates() (*BaseTemplates, error) {
	var err error
	var html map[language.Tag]*html.Template
	var schemaJson map[language.Tag]*text.Template
	var openApiJson map[language.Tag]*text.Template

	html, err = buildBaseTemplatesHTML()
	if err != nil {
		return nil, err
	}
	schemaJson = buildBaseTemplatesJSON()
	openApiJson = buildBaseTemplatesJSON()
	baseTemplates := &BaseTemplates{
		html:        html,
		schemaJSON:  schemaJson,
		openApiJSON: openApiJson,
	}
	return baseTemplates, nil
}

func (b *BaseTemplates) clone() (*BaseTemplates, error) {
	clone := &BaseTemplates{
		html:        make(map[language.Tag]*html.Template, len(b.html)),
		schemaJSON:  make(map[language.Tag]*text.Template, len(b.schemaJSON)),
		openApiJSON: make(map[language.Tag]*text.Template, len(b.openApiJSON)),
	}

	for lang, t := range b.html {
		c, err := t.Clone()
		if err != nil {
			return nil, err
		}
		clone.html[lang] = c
	}

	for lang, t := range b.schemaJSON {
		c, err := t.Clone()
		if err != nil {
			return nil, err
		}
		clone.schemaJSON[lang] = c
	}

	for lang, t := range b.openApiJSON {
		c, err := t.Clone()
		if err != nil {
			return nil, err
		}
		clone.openApiJSON[lang] = c
	}

	return clone, nil
}

func (b *BaseTemplates) MustFork() *BaseTemplates {
	clone, err := b.clone()
	if err != nil {
		panic(fmt.Sprintf("ui: failed to fork base templates: %v", err))
	}
	return clone
}
