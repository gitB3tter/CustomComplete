package transport

import (
	"context"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/weblinking"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// RFC 9457 Problem Details
type Problem struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// Extension

type Validation struct {
	Problem
	InvalidParams errs.InvalidParams `json:"errors,omitempty"`
}

type Conflict struct {
	Problem
	InvalidParams errs.InvalidParams `json:"errors,omitempty"`
}

// identifier
type ProblemType string

const (
	ProblemTypeValidation ProblemType = "validation"
	ProblemTypeConflict   ProblemType = "conflict"
)

var problems = map[ProblemType]struct{}{
	ProblemTypeValidation: {},
}

func ErrJSONTyp(expected, actual JSONType) *errs.CodedError {
	return errs.Type(string(expected), string(actual))
}

func ParseProblemType(s string) (ProblemType, bool) {
	if _, ok := problems[ProblemType(s)]; ok {
		return ProblemType(s), true
	}
	return "", false
}

func RespondProblemValidation(w http.ResponseWriter, r *http.Request, params *errs.InvalidParams, detail string) {
	ctx := r.Context()
	problem, err := ProblemValidation(ctx, params, detail)
	if err != nil {
		InternalErrorGeneric(w, r, fmt.Errorf("Could not marshall validation problem json: %w", err))
		return
	}
	WriteProblemValidation(w, ctx, problem)
}

func ProblemConflict(ctx context.Context, params errs.InvalidParams, detail string) ([]byte, error) {
	T := i18n.TranslateFuncFromContext(ctx)
	problem := Validation{
		Problem: Problem{
			Type:   getProblemTypePath(ProblemTypeConflict).String(),
			Title:  T(i18n.CommonErrorSummaryTitle),
			Detail: detail,
		},
		InvalidParams: params,
	}

	body, err := json.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("Could not marshall validation problem json: %w", err)
	}
	return body, nil
}

func ProblemValidation(ctx context.Context, params *errs.InvalidParams, detail string) ([]byte, error) {
	T := i18n.TranslateFuncFromContext(ctx)
	problem := Validation{
		Problem: Problem{
			Type:   getProblemTypePath(ProblemTypeValidation).String(),
			Title:  T(i18n.CommonErrorSummaryTitle),
			Detail: detail,
		},
		InvalidParams: *params,
	}

	body, err := json.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("Could not marshall validation problem json: %w", err)
	}
	return body, nil
}

func WriteProblemValidation(w http.ResponseWriter, ctx context.Context, problem []byte) {

	SetErrorHeaders(w.Header())

	w.Header().Add("Link", GetProblemDescribedBy(ProblemTypeValidation).String())
	w.Header().Set("Content-Type", mediatypes.ProblemJSON.String())
	w.WriteHeader(http.StatusUnprocessableEntity)
	Write(ctx, w, problem)
}
func RespondProblemGeneric(w http.ResponseWriter, r *http.Request, detail string, statuscode int) {
	output, err := ProblemGeneric(detail)
	if err != nil {
		InternalErrorGeneric(w, r, err)
		return
	}
	SetErrorHeaders(w.Header())

	w.Header().Set("Content-Type", mediatypes.ProblemJSON.String())
	w.WriteHeader(statuscode)
	Write(r.Context(), w, output)
}

func ProblemGeneric(detail string) ([]byte, error) {
	problem := Problem{
		Detail: detail,
	}
	body, err := json.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("Could not marshall problem json. detail: %s, error: %w", problem.Detail, err)
	}
	return body, nil
}

func getProblemTypePath(typ ProblemType) *url.URL {
	return (&url.URL{}).JoinPath("/", "docs", "problems", string(typ))
}

func GetProblemDescribedBy(typ ProblemType) (links weblinking.Links) {
	switch typ {
	case ProblemTypeValidation:
		link := weblinking.New(weblinking.DescribedBy, getProblemTypePath(ProblemTypeValidation))
		links = append(links, link.WithType(mediatypes.HTML), link.WithType(mediatypes.SchemaJSON))
	}
	return links
}
