package ui

import (
	"bytes"
	"customcomplete/internal/i18n"
	"encoding/json"
	"fmt"
	"log"
	"text/template"

	"golang.org/x/text/language"
)

func buildBaseTemplatesJSON() map[language.Tag]*template.Template {

	baseTemplates := make(map[language.Tag]*template.Template)

	for _, langTag := range i18n.SupportedLanguages {
		baseTemplates[langTag] = template.New("").Funcs(
			template.FuncMap{
				"T": func(key i18n.Key, args ...any) string {
					str := i18n.Translate(langTag, key, args...)
					b, err := json.Marshal(str)
					if err != nil || len(b) < 2 {
						return str
					}
					//strip leading and trailing quotes since they are already present in the template file to make it valid JSON
					return string(b[1 : len(b)-1])
				},
			})
	}

	return baseTemplates
}

func prebakeTemplatesJSON(baseTemplates map[language.Tag]*template.Template, templateStrings []string, data map[language.Tag]any) (map[language.Tag][]byte, error) {

	var err error
	var clone *template.Template

	out := make(map[language.Tag][]byte)

	for lang, baseTemplate := range baseTemplates {
		clone, err = baseTemplate.Clone()
		if err != nil {
			return nil, err
		}

		for _, tmplContent := range templateStrings {
			clone, err = clone.Parse(tmplContent)
			if err != nil {
				return nil, fmt.Errorf("failed to parse template: %v", err)
			}
		}

		var buf bytes.Buffer

		if err := clone.Execute(&buf, data[lang]); err != nil {
			return nil, fmt.Errorf("failed to execute template: %v", err)
		}

		if !json.Valid(buf.Bytes()) {
			log.Print(buf.String())
			return nil, fmt.Errorf("template Output is not valid json: %v", err)
		}
		out[lang] = buf.Bytes()
	}

	return out, nil
}
