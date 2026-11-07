package errs

import (
	"fmt"
	"slices"
	"strings"
)

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

	//resource-level
	CodeNotFound Code = "notFound"

	//params
	ParamLimit     param = "limit"
	ParamExpected  param = "expected"
	ParamActual    param = "actual"
	ParamPattern   param = "pattern"
	ParamCanonical param = "canonical"
	ParamRule      param = "rule"
	ParamEncoding  param = "encoding"

	//rules
	RuleNotBlank = "notBlank"
	RuleReserved = "reserved"
)

type Code string
type param string
type Params map[param]any

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
		b.WriteString(string(k))
		if p[k] != "" {
			fmt.Fprintf(&b, "=%v", p[k])
		}
	}
	return b.String()
}

// type of error
type CodedError struct {
	c Code
	p Params
}

func (e *CodedError) Error() string {
	if len(e.p) > 0 {
		return string(e.c) + "{" + e.p.String() + "}"
	}
	return string(e.c)
}

func (e *CodedError) Is(target error) bool {
	x, ok := target.(interface{ Code() Code })
	if !ok {
		return false
	}
	return x.Code() == e.c
}

func (e *CodedError) Code() Code {
	return e.c
}

func (e *CodedError) Params() Params {
	return e.p
}
