package errs

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/schema"
	"customcomplete/internal/weblinking"

	"golang.org/x/text/language"
)

var (
	SchemaPath weblinking.Path = weblinking.DocsPath.Join("problems", "extensions", "errors")

	allCodes = Codes{
		CodeRequired,
		CodePattern,
		CodeMinBytes,
		CodeMaxBytes,
		CodeMinCodePoints,
		CodeMaxCodePoints,
		CodeMinWords,
		CodeMaxWords,
		CodeReadOnly,
		CodeAdditionalProperties,
		CodeType,
		CodeInvalid,
		CodeNotFound,
		CodeConflict,
		CodeEnum,
	}

	allParams = []param{
		ParamLimit,
		ParamExpected,
		ParamActual,
		ParamPattern,
		ParamCanonical,
		ParamRule,
		ParamEncoding,
	}

	allRules = []string{RuleNotBlank}
)

func Schema(langTag language.Tag) schema.Root {
	return FieldError{}.Schema(langTag)
}

func (FieldError) Schema(langTag language.Tag) schema.Root {
	T := i18n.Func(langTag)

	required := map[string]schema.Schema{
		"name": schema.New(schema.String).
			WithMeta(T(i18n.DocsProblemInvalidParamNameTitle), T(i18n.DocsProblemInvalidParamNameDescription)),
		"code": schema.New(schema.String).
			WithMeta(T(i18n.DocsProblemInvalidParamCodeTitle), T(i18n.DocsProblemInvalidParamCodeDescription)).
			WithEnum(allCodes.toAny()...),
		"reason": schema.New(schema.String).
			WithMeta(T(i18n.DocsProblemInvalidParamReasonTitle), T(i18n.DocsProblemInvalidParamReasonDescription)),
	}

	optional := map[string]schema.Schema{
		"params": Params{}.Schema(langTag),
	}

	return schema.NewRoot(
		SchemaPath,
		required,
	).WithTitle(T(i18n.NavigationDocsProblemInvalidParams)).
		WithOptional(optional)
}

func (Params) Schema(langTag language.Tag) schema.Schema {
	T := i18n.Func(langTag)

	return schema.New(schema.Object).
		WithMeta(
			T(i18n.DocsProblemInvalidParamParamsTitle),
			T(i18n.DocsProblemInvalidParamParamsDescription),
		).
		WithProperties(
			map[string]schema.Schema{
				string(ParamLimit): schema.
					New(schema.Integer).
					WithMinimum(0),
				string(ParamActual): schema.
					New(schema.Integer, schema.String).
					WithMinimum(0),
				string(ParamExpected): schema.
					New(schema.String, schema.Array),
				string(ParamPattern): schema.
					New(schema.String).
					WithFormat("regex"),
				string(ParamCanonical): schema.
					New(schema.String),
				string(ParamRule): schema.
					New(schema.String).
					WithEnum(RuleNotBlank),
				string(ParamEncoding): schema.
					New(schema.String).
					WithEnum("utf-8"),
			},
		)
}
