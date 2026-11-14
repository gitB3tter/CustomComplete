package errs

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type param string
type Params map[param]any

type Code string

type Codes []Code

func (codes Codes) toAny() (out []any) {
	out = make([]any, 0, len(codes))
	for _, code := range codes {
		out = append(out, code)
	}
	return
}

func (p Params) String() string {
	keys := make([]param, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		switch v := p[k].(type) {
		case string:
			fmt.Fprintf(&b, "%s=%q", k, v)
		default:
			fmt.Fprintf(&b, "%s=%v", k, v)
		}
	}
	return b.String()
}

const (
	//codes
	//field-level (json-schema/openapi-inspired/adjacent)
	CodeRequired             Code = "required"
	CodeMinBytes             Code = "minBytes"
	CodeMaxBytes             Code = "maxBytes"
	CodeMinCodePoints        Code = "minCodePoints"
	CodeMaxCodePoints        Code = "maxCodePoints"
	CodeMinWords             Code = "minWords"
	CodeMaxWords             Code = "maxWords"
	CodePattern              Code = "pattern"
	CodeReadOnly             Code = "readOnly"
	CodeAdditionalProperties Code = "additionalProperties"
	CodeType                 Code = "type"
	CodeConflict             Code = "conflict"
	CodeEnum                 Code = "enum"
	//catch-all
	CodeInvalid Code = "invalid"

	//extension

	CodeNotFound Code = "notFound"

	//params
	Limit        param = "limit"
	Expected     param = "expected"
	Actual       param = "actual"
	ParamPattern param = "pattern"
	Canonical    param = "canonical"
	Rule         param = "rule"
	Encoding     param = "encoding"

	//rules
	NotBlank = "notBlank"
)

// type of error
type Coded struct {
	c Code
	p Params
}

func (errCoded *Coded) Error() string {
	if len(errCoded.p) > 0 {
		return string(errCoded.c) + "{" + errCoded.p.String() + "}"
	}
	return string(errCoded.c)
}

func (errCoded *Coded) Code() Code {
	return errCoded.c
}

func (errCoded *Coded) Params() Params {
	return errCoded.p
}

func ParamOr[T any](errCoded Coded, key param, fallback T) T {
	if val, ok := errCoded.p[key].(T); ok {
		return val
	}
	return fallback
}

func HasCode(err error, code Code) bool {
	coded, ok := errors.AsType[*Coded](err)
	return ok && coded.Code() == code
}
