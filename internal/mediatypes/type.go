package mediatypes

import (
	"errors"
	"fmt"
	"mime"
	"strings"
)

// concrete Media Type without wildcards
type Type struct {
	typ     string
	subtype string
	//canonical
	params parameters
}

type Types []Type

var (
	JSON = must(ParseMediaType("application/json"))

	OpenAPIJSON = must(ParseMediaType("application/vnd.oai.openapi+json"))

	LinksetJSON = must(ParseMediaType("application/linkset+json"))

	MergePatchJSON = must(ParseMediaType("application/merge-patch+json"))

	// rfc 9457
	ProblemJSON = must(ParseMediaType("application/problem+json"))

	SchemaJSON = must(ParseMediaType("application/schema+json"))

	// text/html (regarding charset: utf-8: see html living standard § 17.1 text/html)
	HTML = must(ParseMediaType("text/html;charset=utf-8"))

	Linkset = must(ParseMediaType("application/linkset"))

	FormUrlEncoded = must(ParseMediaType("application/x-www-form-urlencoded"))

	MultipartForm = must(ParseMediaType("multipart/form-data"))

	TextEventStream = must(ParseMediaType("text/event-stream"))

	TextPlain = must(ParseMediaType("text/plain;charset=utf-8"))

	URLEncoded = must(ParseMediaType("application/x-www-form-urlencoded"))
)

const (
	typeMultipart = "multipart"
)

func ParseMediaType(field string) (mediaType Type, err error) {
	mtString, params, err := mime.ParseMediaType(field)
	if err != nil {
		return Type{}, err
	}
	var ok bool
	mediaType.typ, mediaType.subtype, ok = strings.Cut(mtString, "/")
	if !ok {
		return Type{}, fmt.Errorf("no mediatype: %q is not type/subtype", field)
	}
	if mediaType.typ == "*" || mediaType.subtype == "*" {
		return Type{}, errors.New("no mediatype: contains wildcard")
	}

	mediaType.params = canonizeParams(params)

	return mediaType, nil
}

func must(mediaType Type, err error) Type {
	if err != nil {
		panic(err)
	}
	return mediaType
}

// e.g. for declaring available representations in 406 response
func (types Types) String() string {
	entries := make([]string, 0, len(types))
	for _, mediaType := range types {
		entries = append(entries, mediaType.String())
	}
	return strings.Join(entries, ",")
}

func (mediaType Type) String() string {
	if mediaType.IsZero() {
		return ""
	}
	if mediaType.params == "" {
		return mediaType.typ + "/" + mediaType.subtype
	}
	return mediaType.typ + "/" + mediaType.subtype + "; " + string(mediaType.params)
}

func (mediaType Type) syntax() string {
	if idx := strings.LastIndex(mediaType.subtype, "+"); idx != -1 {
		return mediaType.subtype[idx+1:]
	} else {
		return mediaType.subtype
	}
}

func (mediaType Type) SharesSyntax(other Type) bool {
	return mediaType.syntax() == other.syntax()
}

func (mediaType Type) IsZero() bool {
	return mediaType.typ == "" || mediaType.subtype == ""
}
func (mediaType Type) MarshalText() ([]byte, error) {
	if mediaType.IsZero() {
		return nil, nil
	}
	return []byte(mediaType.String()), nil
}
