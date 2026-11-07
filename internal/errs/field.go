package errs

import (
	"strings"
)

// cause + field/param/name
type FieldError struct {
	Name string
	*CodedError
}

func FieldFromCoded(name string, coded *CodedError) *FieldError {
	return &FieldError{
		Name:       name,
		CodedError: coded,
	}
}

func (e *FieldError) Error() string {
	return "field=" + e.Name + " error=" + e.Coded().Error()
}

func (e *FieldError) Unwrap() error {
	return e.CodedError
}

func (e *FieldError) Coded() *CodedError {
	return e.CodedError
}

type FieldErrors struct {
	errs []*FieldError
}

func (e *FieldErrors) Error() string {
	var b strings.Builder
	b.WriteString("field_errors=[")

	for i := range e.errs {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteByte('{')
		b.WriteString(e.errs[i].Error())
		b.WriteByte('}')
	}

	b.WriteByte(']')
	return b.String()
}

func (e *FieldErrors) Errors() []*FieldError {
	return e.errs
}

func (e *FieldErrors) Add(name string, err *CodedError) {
	e.errs = append(e.errs, &FieldError{Name: name, CodedError: err})
}

func (e *FieldErrors) Append(errs ...*FieldError) {
	e.errs = append(e.errs, errs...)
}

func (e *FieldErrors) OK() bool {
	return len(e.errs) == 0
}
