package ui

import (
	"customcomplete/internal/mediatypes"
	"fmt"

	"golang.org/x/text/language"
)

type Prebaked map[language.Tag][]byte

func (baseTemplates *BaseTemplates) Prebake(typ mediatypes.Type, templates []string, data map[language.Tag]any) (Prebaked, error) {

	switch typ {
	case mediatypes.HTML:
		return baseTemplates.prebakeHTML(templates, data)
	case mediatypes.SchemaJSON:
		return prebakeTemplatesJSON(baseTemplates.schemaJSON, templates, data)
	case mediatypes.OpenAPIJSON:
		return prebakeTemplatesJSON(baseTemplates.openApiJSON, templates, data)
	}
	return map[language.Tag][]byte{}, fmt.Errorf("prebaking for mediaType %s is not implemented", typ.String())
}
