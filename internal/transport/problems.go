package transport

import (
	"context"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/schema"
	"customcomplete/internal/weblinking"
	"encoding/json"
	"fmt"

	"golang.org/x/text/language"
)

// RFC 9457 Problem Details
type Problem struct {
	Type   weblinking.Path `json:"type,omitempty"`
	Title  string          `json:"title,omitempty"`
	Detail string          `json:"detail,omitempty"`
}

// Extension

type Validation struct {
	Problem
	InvalidParams errs.FieldErrors `json:"errors,omitempty"`
}

type Conflict struct {
	Problem
	InvalidParams errs.FieldErrors `json:"errors,omitempty"`
}

var (
	ProblemErrorsPath  weblinking.Path = weblinking.DocsPath.Join("problems", "errors")
	ProblemGenericPath weblinking.Path = weblinking.DocsPath.Join("problems", "generic")
)

func ProblemErrorSchema(langTag language.Tag) schema.Root {
	T := i18n.Func(langTag)

	required := map[string]schema.Schema{
		"type": schema.New(schema.String).
			WithConst(ProblemErrorsPath),
		"title": schema.New(schema.String).
			WithDescription(T(i18n.DocsProblemTitleDescription)),
		"errors": schema.New(schema.Array).
			WithItemsRef(errs.SchemaPath).
			WithDescription(T(i18n.DocsProblemInvalidParamsDescription)),
	}

	return schema.NewRoot(
		ProblemErrorsPath,
		required,
	).WithOptional(map[string]schema.Schema{
		"detail": schema.New(schema.String),
	})
}

func ProblemGenericSchema(langTag language.Tag) schema.Root {

	required := map[string]schema.Schema{
		"detail": schema.New(schema.String),
	}

	return schema.NewRoot(
		ProblemGenericPath,
		required,
	)
}

func ProblemFieldError(ctx context.Context, invalidParams errs.FieldErrors, detail string) ([]byte, error) {
	T := i18n.TranslateFuncFromContext(ctx)
	problem := Validation{
		Problem: Problem{
			Type:   ProblemErrorsPath,
			Title:  T(i18n.CommonErrorSummaryTitle),
			Detail: detail,
		},
		InvalidParams: invalidParams,
	}

	body, err := json.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("Could not marshall field error problem json: %w", err)
	}
	return body, nil
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
