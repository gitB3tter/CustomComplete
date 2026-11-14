package errs

import (
	"strings"
)

type Field struct {
	Name   string `json:"name"`
	Code   Code   `json:"code"`
	Params Params `json:"params,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func FieldFromCoded(name string, errCoded Coded, reason string) Field {
	return Field{
		Name:   name,
		Code:   errCoded.Code(),
		Params: errCoded.Params(),
		Reason: reason,
	}
}

func (errField *Field) String() string {
	var b strings.Builder
	b.WriteString(errField.Name)
	b.WriteString(": ")
	b.WriteString(errField.Reason)
	b.WriteString(" [")
	b.WriteString(string(errField.Code))
	if len(errField.Params) > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(errField.Params.String())
	b.WriteByte(']')
	return b.String()
}

func (errField *Field) Error() string {
	return errField.String()
}

type Fields []Field

func (errsField *Fields) AddCoded(name string, errCoded Coded, reason string) {
	*errsField = append(*errsField, FieldFromCoded(name, errCoded, reason))
}

func (errsField *Fields) String() string {
	return errsField.Error()
}
func (errsField *Fields) Error() string {
	var b strings.Builder
	for i, errField := range *errsField {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("- ")
		b.WriteString(errField.String())
	}
	return b.String()
}
