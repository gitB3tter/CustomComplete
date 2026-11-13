package constraints

import (
	"customcomplete/internal/errs"
	"customcomplete/internal/schema"
)

const (
	Sha256HexLength  int    = 64
	Sha256HexPattern string = `^[0-9a-f]{64}$`
)

func (NotBlank) Schema() schema.Schema {
	return schema.Schema{
		Type:          schema.String,
		MinCodepoints: 1,
		Pattern:       `\S`,
		XConstraints: schema.Constraints{
			Rules: []string{errs.RuleNotBlank},
		},
	}
}

func (Label) Schema() schema.Schema {
	s := NotBlank("").Schema()
	s.MaxCodepoints = ShortStringMaxCodePoints
	return s
}

func (SafeFilename) Schema() schema.Schema {
	s := NotBlank("").Schema()
	s.MinCodepoints = safeFileNameMinCodePoints
	s.MaxCodepoints = safeFileNameMaxBytes
	s.Pattern = SafeFilenamePattern
	return s
}

func (Sha256Hex) Schema() schema.Schema {
	return schema.Schema{
		Type:          schema.String,
		Pattern:       Sha256HexPattern,
		MinCodepoints: Sha256HexLength,
		MaxCodepoints: Sha256HexLength,
	}
}

func (FileContents) Schema() schema.Schema {
	return schema.Schema{
		Type: schema.String,
		XConstraints: schema.Constraints{
			MaxBytes: MaxFileSize,
		},
	}
}
