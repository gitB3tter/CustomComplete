package errs

func NotFound() *CodedError {
	return &CodedError{
		c: CodeNotFound,
	}
}
func Required() *CodedError {
	return &CodedError{
		c: CodeRequired,
	}
}
func Blank() *CodedError {
	return &CodedError{
		c: CodeRequired,
		p: Params{ParamRule: RuleNotBlank},
	}
}

func MaxBytes(limit int64) *CodedError {
	return &CodedError{
		c: CodeMaxBytes,
		p: Params{ParamLimit: limit},
	}
}

func MinCodePoints(limit int, actual int) *CodedError {
	return &CodedError{
		c: CodeMinCodePoints,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func MaxCodePoints(limit int, actual int) *CodedError {
	return &CodedError{
		c: CodeMaxCodePoints,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func MinWords(limit int, actual int) *CodedError {
	return &CodedError{
		c: CodeMinWords,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func MaxWords(limit int, actual int) *CodedError {
	return &CodedError{
		c: CodeMaxWords,
		p: Params{ParamLimit: limit, ParamActual: actual},
	}
}

func Pattern(pattern string) *CodedError {
	return &CodedError{
		c: CodePattern,
		p: Params{ParamPattern: pattern},
	}
}

func PatternWithCanonical(pattern string, canonical string) *CodedError {
	return &CodedError{
		c: CodePattern,
		p: Params{ParamPattern: pattern, ParamCanonical: canonical},
	}
}

func ReadOnly() *CodedError {
	return &CodedError{
		c: CodeReadOnly,
	}
}

func Type(expected, actual string) *CodedError {
	return &CodedError{
		c: CodeType,
		p: Params{ParamExpected: expected, ParamActual: actual},
	}
}

func Enum(expected []any, actual any) *CodedError {
	return &CodedError{
		c: CodeEnum,
		p: Params{ParamExpected: expected, ParamActual: actual},
	}
}

func Invalid(params Params) *CodedError {
	return &CodedError{
		c: CodeInvalid,
		p: params,
	}
}

func InvalidUTF8() *CodedError {
	return &CodedError{
		c: CodeInvalid,
		p: Params{
			ParamEncoding: "utf-8",
		},
	}
}

func Unexpected() *CodedError {
	return &CodedError{
		c: CodeAdditionalProperties,
	}
}

func Conflict(expected, actual any) *CodedError {
	return &CodedError{
		c: CodeConflict,
		p: Params{ParamExpected: expected, ParamActual: actual},
	}
}

func FieldRequired(name string) *FieldError {
	return &FieldError{
		Name:       name,
		CodedError: Required(),
	}
}

func FieldInvalid(name string) *FieldError {
	return &FieldError{
		Name:       name,
		CodedError: Invalid(Params{}),
	}
}

func Join(errs ...*FieldError) *FieldErrors {
	if len(errs) == 0 {
		return nil
	}

	out := make([]*FieldError, 0, len(errs))
	for _, e := range errs {
		if e != nil {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return &FieldErrors{errs: out}
}
