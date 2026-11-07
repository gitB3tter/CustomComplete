package i18n

import (
	"regexp"
	"slices"
	"testing"
)

var fmtVerbRe = regexp.MustCompile(`%[sdv%]`)

func TestCompletenessAndFormatVerbs(t *testing.T) {
	for k, en := range translationsEn {
		if en == "" {
			t.Errorf("en: key %q empty", k)
		}
		de, ok := translationsDe[k]
		if !ok {
			t.Errorf("de: key %q missing", k)
			continue
		}
		if de == "" {
			t.Errorf("de: key %q empty", k)
		}

		VerbsEn := fmtVerbRe.FindAllString(en, -1)
		VerbsDe := fmtVerbRe.FindAllString(de, -1)
		if !slices.Equal(VerbsEn, VerbsDe) {
			t.Errorf("key %q format verbs mismatch: en=%v (%q) vs de=%v (%q)", k, VerbsEn, en, VerbsDe, de)
		}
	}

	for k := range translationsDe {
		if _, ok := translationsEn[k]; !ok {
			t.Errorf("en: key %q missing", k)
		}
	}
}
