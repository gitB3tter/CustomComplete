package errs

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"golang.org/x/text/language"
)

func TestFieldErrorSchema(t *testing.T) {
	s := Schema(language.English)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	t.Logf("Generated schema:\n%s", string(b))

	for _, req := range []string{"name", "code", "reason"} {
		if !slices.Contains(s.Required, req) {
			t.Errorf("schema.Required missing field: %q", req)
		}
	}

	codeProp, ok := s.Properties["code"]
	if !ok {
		t.Fatalf(`schema does not contain "code" property`)
	}
	for _, code := range allCodes {
		if !slices.Contains(codeProp.Enum, any(code)) {
			t.Errorf("code enum missing code: %q", code)
		}
	}

	if !slices.Contains(slices.Collect(maps.Keys(s.Properties)), "params") {
		t.Fatalf(`schema does not contain "params"`)
	}

	paramsProp := s.Properties["params"]
	for _, p := range allParams {
		if _, ok := paramsProp.Properties[string(p)]; !ok {
			t.Errorf("params schema missing param: %q", p)
		}
	}

	ruleProp, ok := paramsProp.Properties[string(Rule)]
	if !ok {
		t.Fatalf(`params schema does not contain %q property`, Rule)
	}
	for _, r := range allRules {
		if !slices.Contains(ruleProp.Enum, any(r)) {
			t.Errorf("rule enum missing rule: %q", r)
		}
	}
}
