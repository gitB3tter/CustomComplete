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
		Limit,
		Expected,
		Actual,
		ParamPattern,
		Canonical,
		Rule,
		Encoding,
	}

	allRules = []string{NotBlank}
)

func Schema(langTag language.Tag) schema.Root {
	return Field{}.Schema(langTag)
}

func (Field) Schema(langTag language.Tag) schema.Root {
	T := i18n.Func(langTag)

	required := map[string]schema.Schema{
		"name": schema.New(schema.String).
			WithMeta(T(i18n.DocsProblemsFieldErrorNameTitle), T(i18n.DocsProblemsFieldErrorNameDescription)),
		"code": schema.New(schema.String).
			WithMeta(T(i18n.DocsProblemsFieldErrorCodeTitle), T(i18n.DocsProblemsFieldErrorCodeDescription)).
			WithEnum(allCodes.toAny()...),
		"reason": schema.New(schema.String).
			WithMeta(T(i18n.DocsProblemsFieldErrorReasonTitle), T(i18n.DocsProblemsFieldErrorReasonDescription)),
	}

	optional := map[string]schema.Schema{
		"params": Params{}.Schema(langTag),
	}

	return schema.NewRoot(
		SchemaPath,
		required,
	).WithTitle(T(i18n.NavigationDocsProblemsError)).
		WithOptional(optional)
}

func (Params) Schema(langTag language.Tag) schema.Schema {
	T := i18n.Func(langTag)

	return schema.New(schema.Object).
		WithMeta(
			T(i18n.DocsProblemsFieldErrorParamsTitle),
			T(i18n.DocsProblemsFieldErrorParamsDescription),
		).
		WithProperties(
			map[string]schema.Schema{
				string(Limit): schema.
					New(schema.Integer).
					WithMinimum(0),
				string(Actual): schema.
					New(schema.Integer, schema.String).
					WithMinimum(0),
				string(Expected): schema.
					New(schema.String, schema.Array),
				string(ParamPattern): schema.
					New(schema.String).
					WithFormat("regex"),
				string(Canonical): schema.
					New(schema.String),
				string(Rule): schema.
					New(schema.String).
					WithEnum(NotBlank),
				string(Encoding): schema.
					New(schema.String).
					WithEnum("utf-8"),
			},
		)
}
