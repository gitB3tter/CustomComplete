package errs

import (
	"strings"
)

type FieldError struct {
	Name   string `json:"name"`
	Code   Code   `json:"code"`
	Params Params `json:"params,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func InvalidParamFromCodedError(name string, errCoded CodedError, reason string) FieldError {
	return FieldError{
		Name:   name,
		Code:   errCoded.Code(),
		Params: errCoded.Params(),
		Reason: reason,
	}
}

func (errField *FieldError) String() string {
	var b strings.Builder
	b.WriteString(errField.Name)
	b.WriteString(": ")
	b.WriteString(errField.Reason)
	b.WriteString(" [")
	b.WriteString(string(errField.Code))
	b.WriteByte(' ')
	b.WriteString(errField.Params.String())
	b.WriteByte(']')
	return b.String()
}

func (errField *FieldError) Error() string {
	return errField.String()
}

type FieldErrors []FieldError

func (errsField *FieldErrors) AddCoded(name string, errCoded CodedError, reason string) {
	*errsField = append(*errsField, InvalidParamFromCodedError(name, errCoded, reason))
}

func (errsField *FieldErrors) String() string {
	return errsField.Error()
}
func (errsField *FieldErrors) Error() string {
	var b strings.Builder
	for i, iParam := range *errsField {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("- ")
		b.WriteString(iParam.String())
	}
	return b.String()
}
