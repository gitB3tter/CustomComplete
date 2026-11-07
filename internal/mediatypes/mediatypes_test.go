package mediatypes

import (
	"testing"
)

func TestNegotiateRFC9110(t *testing.T) {
	supported := Types{HTML, JSON, LinksetJSON}

	tests := []struct {
		name        string
		accept      []string
		support     Types
		expectType  Type
		expectMatch bool
	}{
		{
			name:        "empty Accept",
			accept:      nil,
			support:     supported,
			expectType:  supported[0],
			expectMatch: true,
		},
		{
			name:        "wildcard + explicit exclusion",
			accept:      []string{"*/*, application/json;q=0"},
			support:     Types{JSON, HTML},
			expectType:  HTML,
			expectMatch: true,
		},
		{
			name:        "supported type excluded",
			accept:      []string{"text/html;q=0.000"},
			support:     Types{HTML},
			expectType:  Type{},
			expectMatch: false,
		},
		{
			name:        "specific > wildcard when both apply to type",
			accept:      []string{"*/*;q=0.9, application/json;q=0.4"},
			support:     Types{JSON, HTML},
			expectType:  HTML,
			expectMatch: true,
		},
		{
			name:        "client order as last tiebreaker (specificity irrelevent between same weight types)",
			accept:      []string{"application/json, text/html; charset=utf-8"},
			support:     Types{HTML, JSON},
			expectType:  JSON,
			expectMatch: true,
		},
		{
			name:        "wildcard w. higher weight > exact type",
			accept:      []string{"text/*;q=0.9, text/html;q=0.5"},
			support:     Types{HTML, TextPlain},
			expectType:  TextPlain,
			expectMatch: true,
		},
		{
			name:        "accept-ext",
			accept:      []string{"text/html;q=0.4, application/linkset+json;q=0.9;ext=custom"},
			support:     supported,
			expectType:  LinksetJSON,
			expectMatch: true,
		},
		{
			name:        "unknown media type parameter ",
			accept:      []string{"text/html;q=0.4, application/json;charset=utf-8;q=0.9"},
			support:     supported,
			expectType:  HTML,
			expectMatch: true,
		},
		{
			name:        "multiple Accept header lines",
			accept:      []string{"text/plain;q=0.5", "application/json;q=0.9"},
			support:     supported,
			expectType:  JSON,
			expectMatch: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := ParseAccept(tc.accept)
			gotType, gotMatch := tc.support.Negotiate(parsed)
			if gotMatch != tc.expectMatch {
				t.Fatalf("Negotiate(%v) match = %v, want %v (gotType=%v)", tc.accept, gotMatch, tc.expectMatch, gotType)
			}
			if gotMatch && gotType != tc.expectType {
				t.Errorf("Negotiate(%v) type = %v, want %v", tc.accept, gotType, tc.expectType)
			}
		})
	}
}
