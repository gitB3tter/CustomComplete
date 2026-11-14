package errs

import "customcomplete/internal/schema"

var (
	errNotFound             = Coded{c: CodeNotFound}
	errRequired             = Coded{c: CodeRequired}
	errReadOnly             = Coded{c: CodeReadOnly}
	errAdditionalProperties = Coded{c: CodeAdditionalProperties}
)

func NotFound() Coded {
	return errNotFound
}
func Required() Coded {
	return errRequired
}

func ReadOnly() Coded {
	return errReadOnly
}

func AdditionalProperties() Coded {
	return errAdditionalProperties
}

func Blank() Coded {
	return Coded{
		c: CodeRequired,
		p: Params{Rule: NotBlank},
	}
}

func MaxBytes(limit int64) Coded {
	return Coded{
		c: CodeMaxBytes,
		p: Params{Limit: limit},
	}
}

func MinCodePoints(limit int, actual int) Coded {
	return Coded{
		c: CodeMinCodePoints,
		p: Params{Limit: limit, Actual: actual},
	}
}

func MaxCodePoints(limit int, actual int) Coded {
	return Coded{
		c: CodeMaxCodePoints,
		p: Params{Limit: limit, Actual: actual},
	}
}

func MinWords(limit int, actual int) Coded {
	return Coded{
		c: CodeMinWords,
		p: Params{Limit: limit, Actual: actual},
	}
}

func MaxWords(limit int, actual int) Coded {
	return Coded{
		c: CodeMaxWords,
		p: Params{Limit: limit, Actual: actual},
	}
}

func Pattern(pattern string) Coded {
	return Coded{
		c: CodePattern,
		p: Params{ParamPattern: pattern},
	}
}

func PatternWithCanonical(pattern string, canonical string) Coded {
	return Coded{
		c: CodePattern,
		p: Params{ParamPattern: pattern, Canonical: canonical},
	}
}

func Type(expected, actual schema.Type) Coded {
	return Coded{
		c: CodeType,
		p: Params{Expected: expected, Actual: actual},
	}
}

func Enum(expected []any, actual any) Coded {
	return Coded{
		c: CodeEnum,
		p: Params{Expected: expected, Actual: actual},
	}
}

func Invalid(params Params) Coded {
	return Coded{
		c: CodeInvalid,
		p: params,
	}
}

func InvalidUTF8() Coded {
	return Coded{
		c: CodeInvalid,
		p: Params{
			Encoding: "utf-8",
		},
	}
}

func Conflict(expected, actual any) Coded {
	return Coded{
		c: CodeConflict,
		p: Params{Expected: expected, Actual: actual},
	}
}
