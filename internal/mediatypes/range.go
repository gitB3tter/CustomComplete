package mediatypes

import (
	"fmt"
	"mime"
	"strings"
)

// (like) media-type (but) possibly with wildcards (RFC 9110 §12.5.1)
type mRange struct {
	typ     string
	subtype string
	params  map[string]string
}

// Returns the Specificity of the Media-Range based on Type and Subtype as a number
func (ma mRange) specificity() int {
	if ma.typ == "*" {
		return 0
	}
	if ma.subtype == "*" {
		return 1
	}
	return 2
}

// parses a mediarange from passed field
func parseMediaRange(field string) (mediaR mRange, err error) {
	//lowercases mediaRange, parameter keys, preserves parameter values
	field, parameters, err := mime.ParseMediaType(field)
	if err != nil {
		return mRange{}, err
	}
	typ, subtype, ok := strings.Cut(field, "/")
	if !ok || typ == "" || subtype == "" {
		return mRange{}, fmt.Errorf("no mediarange: %q is not type/subtype", field)
	}

	return mRange{
		typ:     typ,
		subtype: subtype,
		params:  normalizeParams(parameters),
	}, nil
}
