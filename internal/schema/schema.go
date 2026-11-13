package schema

import (
	"customcomplete/internal/weblinking"
	"maps"
	"reflect"
	"slices"
)

type schemaURI string

const (
	uri schemaURI = "https://json-schema.org/draft/2020-12/schema"
)

func NewRoot(path weblinking.Path, required map[string]Schema) Root {
	return Root{
		SchemaURI:  uri,
		ID:         path,
		Type:       Object,
		Required:   slices.Sorted(maps.Keys(required)),
		Properties: required,
	}
}

type Root struct {
	SchemaURI            schemaURI         `json:"$schema"`
	ID                   weblinking.Path   `json:"$id"`
	Title                string            `json:"title,omitzero"`
	Type                 Type              `json:"type"`
	Required             []string          `json:"required,omitzero"`
	AdditionalProperties bool              `json:"additionalProperties"`
	Properties           map[string]Schema `json:"properties"`
}

func (root Root) WithTitle(title string) Root {
	root.Title = title
	return root
}

func (root Root) WithAdditionalProperties() Root {
	root.AdditionalProperties = true
	return root
}

func (root Root) WithOptional(optional map[string]Schema) Root {
	if root.Properties == nil {
		root.Properties = make(map[string]Schema)
	}
	maps.Copy(root.Properties, optional)
	return root
}

func New(types ...Type) Schema {
	s := Schema{}
	if slices.Contains(types, Object) {
		s.AdditionalProperties = new(false)
	}
	if len(types) == 1 {
		s.Type = types[0]
		return s
	}
	if len(types) > 1 {
		s.Type = types
	}
	return s
}

func From[T any](_ ...T) Schema {
	var zero T
	switch any(zero).(type) {
	case string:
		return Schema{Type: String}
	case bool:
		return Schema{Type: Boolean}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr:
		return Schema{Type: Integer}
	case float32, float64:
		return Schema{Type: Number}
	case []string, []int, []int64, []any, []byte:
		return Schema{Type: Array}
	case map[string]any, map[string]string:
		return Schema{Type: Object, AdditionalProperties: new(false)}
	default:
		switch reflect.TypeOf(&zero).Elem().Kind() {
		case reflect.Slice, reflect.Array:
			return Schema{Type: Array}
		case reflect.Map, reflect.Struct:
			return Schema{Type: Object, AdditionalProperties: new(false)}
		default:
			return Schema{}
		}
	}
}

type Schema struct {
	Title         string            `json:"title,omitzero"`
	Description   string            `json:"description,omitzero"`
	Type          any               `json:"type,omitzero"`
	Const         any               `json:"const,omitzero"`
	Format        string            `json:"format,omitzero"`
	MinCodepoints int               `json:"minLength,omitzero"`
	MaxCodepoints int               `json:"maxLength,omitzero"`
	Minimum       *int64            `json:"minimum,omitzero"`
	Maximum       *int64            `json:"maximum,omitzero"`
	Pattern       string            `json:"pattern,omitzero"`
	ReadOnly      bool              `json:"readOnly,omitzero"`
	Enum          []any             `json:"enum,omitzero"`
	Examples      []any             `json:"examples,omitzero"`
	Properties    map[string]Schema `json:"properties,omitzero"`
	Items         struct {
		Ref weblinking.Path `json:"$ref"`
	} `json:"items,omitzero"`
	AdditionalProperties *bool `json:"additionalProperties,omitzero"`
	//all codes that are not directly mappable to json keywords
	XConstraints Constraints `json:"x-constraints,omitzero"`
}

func (s Schema) WithTypes(types ...Type) Schema {
	if len(types) == 0 {
		return s
	}

	var all []Type
	switch curr := s.Type.(type) {
	case Type:
		all = []Type{curr}
	case []Type:
		all = slices.Clone(curr)
	}

	for _, t := range types {
		if !slices.Contains(all, t) {
			all = append(all, t)
		}
	}

	if len(all) == 1 {
		s.Type = all[0]
	} else {
		s.Type = all
	}
	return s
}

func (s Schema) WithAdditionalProperties() Schema {
	s.AdditionalProperties = new(true)
	return s
}

func (schema Schema) WithMeta(title, description string) Schema {
	schema.Title = title
	return schema.WithDescription(description)
}

func (schema Schema) WithDescription(description string) Schema {
	schema.Description = description
	return schema
}

func (s Schema) WithEnum(values ...any) Schema {
	s.Enum = values
	return s
}

func (s Schema) WithConst(value any) Schema {
	s.Const = value
	return s
}

func (s Schema) WithExamples(examples ...any) Schema {
	s.Examples = examples
	return s
}

func (s Schema) WithMinimum(min int64) Schema {
	s.Minimum = &min
	return s
}

func (s Schema) WithMaximum(max int64) Schema {
	s.Maximum = &max
	return s
}

func (s Schema) WithFormat(format string) Schema {
	s.Format = format
	return s
}

func (s Schema) WithProperties(properties map[string]Schema) Schema {
	if s.AdditionalProperties == nil {
		s.AdditionalProperties = new(false)
	}
	s.Properties = properties
	return s
}

func (s Schema) WithItemsRef(path weblinking.Path) Schema {
	s.Items.Ref = path
	return s
}

func (s Schema) AsReadOnly() Schema {
	s.ReadOnly = true
	return s
}

type Constraints struct {
	MinBytes int64    `json:"minBytes,omitzero"`
	MaxBytes int64    `json:"maxBytes,omitzero"`
	MinWords int      `json:"minWords,omitzero"`
	MaxWords int      `json:"maxWords,omitzero"`
	Encoding string   `json:"encoding,omitzero"`
	Rules    []string `json:"rules,omitzero"`
}
