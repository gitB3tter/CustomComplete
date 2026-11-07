package errs

import (
	"strings"
)

type InvalidParam struct {
	Name   string `json:"name"`
	Code   Code   `json:"code"`
	Params Params `json:"params,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func InvalidParamFromCodedError(name string, e *CodedError, reason string) *InvalidParam {
	return &InvalidParam{
		Name:   name,
		Code:   e.Code(),
		Params: e.Params(),
		Reason: reason,
	}
}

func InvalidParamFromFieldError(e *FieldError, reason string) *InvalidParam {
	return &InvalidParam{
		Name:   e.Name,
		Code:   e.Code(),
		Params: e.Params(),
		Reason: reason,
	}
}

func (iParam *InvalidParam) String() string {
	var b strings.Builder
	b.WriteString(iParam.Name)
	b.WriteString(": ")
	b.WriteString(iParam.Reason)
	b.WriteString(" [")
	b.WriteString(string(iParam.Code))
	b.WriteByte(' ')
	b.WriteString(iParam.Params.String())
	b.WriteByte(']')
	return b.String()
}

func (iParam *InvalidParam) Error() string {
	return iParam.String()
}

type InvalidParams []*InvalidParam

func (iParams *InvalidParams) AddCoded(name string, err *CodedError, reason string) {
	*iParams = append(*iParams, InvalidParamFromCodedError(name, err, reason))
}

func (iParams *InvalidParams) AddField(err *FieldError, reason string) {
	*iParams = append(*iParams, InvalidParamFromFieldError(err, reason))
}

func (iParams *InvalidParams) String() string {
	return iParams.Error()
}
func (iParams *InvalidParams) Error() string {
	var b strings.Builder
	for i, iParam := range *iParams {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("- ")
		b.WriteString(iParam.String())
	}
	return b.String()
}
