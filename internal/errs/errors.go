package errs

import "customcomplete/internal/schema"

var (
	errNotFound             = CodedError{c: CodeNotFound}
	errRequired             = CodedError{c: CodeRequired}
	errReadOnly             = CodedError{c: CodeReadOnly}
	errAdditionalProperties = CodedError{c: CodeAdditionalProperties} // etc.
)

func NotFound() CodedError {
	return errNotFound
}
func Required() CodedError {
	return errRequired
}

func ReadOnly() CodedError {
	return errReadOnly
}

func AdditionalProperties() CodedError {
	return errAdditionalProperties
}

func Blank() CodedError {
	return CodedError{
		c: CodeRequired,
		p: Params{ParamRule: RuleNotBlank},
	}
}

func MaxBytes(limit int64) CodedError {
	return CodedError{
		c: CodeMaxBytes,
		p: Params{ParamLimit: limit},
	}
}

func MinCodePoints(limit int, actual int) CodedError {
	return CodedError{
		c: CodeMinCodePoints,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func MaxCodePoints(limit int, actual int) CodedError {
	return CodedError{
		c: CodeMaxCodePoints,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func MinWords(limit int, actual int) CodedError {
	return CodedError{
		c: CodeMinWords,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func MaxWords(limit int, actual int) CodedError {
	return CodedError{
		c: CodeMaxWords,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func Pattern(pattern string) CodedError {
	return CodedError{
		c: CodePattern,
		p: Params{ParamPattern: pattern},
	}
}

func PatternWithCanonical(pattern string, canonical string) CodedError {
	return CodedError{
		c: CodePattern,
		p: Params{ParamPattern: pattern, ParamCanonical: canonical},
	}
}

func Type(expected, actual schema.Type) CodedError {
	return CodedError{
		c: CodeType,
		p: Params{ParamExpected: expected, ParamActual: actual},
	}
}

func Enum(expected []any, actual any) CodedError {
	return CodedError{
		c: CodeEnum,
		p: Params{ParamExpected: expected, ParamActual: actual},
	}
}

func Invalid(params Params) CodedError {
	return CodedError{
		c: CodeInvalid,
		p: params,
	}
}

func InvalidUTF8() CodedError {
	return CodedError{
		c: CodeInvalid,
		p: Params{
			ParamEncoding: "utf-8",
		},
	}
}

func Conflict(expected, actual any) CodedError {
	return CodedError{
		c: CodeConflict,
		p: Params{ParamExpected: expected, ParamActual: actual},
	}
}
