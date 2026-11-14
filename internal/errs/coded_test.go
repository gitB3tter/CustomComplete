package errs

import "testing"

func TestParamsString(t *testing.T) {
	tests := []struct {
		name   string
		params Params
		expect string
	}{
		{
			name:   "nil params",
			params: nil,
			expect: "",
		},
		{
			name:   "empty params",
			params: Params{},
			expect: "",
		},
		{
			name: "sorted keys with mixed types",
			params: Params{
				Rule:         NotBlank,
				Limit:        10,
				Actual:       0,
				ParamPattern: "^[a-z]+$",
				Expected:     []any{"text", "file"},
			},
			expect: `actual=0, expected=[text file], limit=10, pattern="^[a-z]+$", rule="notBlank"`,
		},
		{
			name: "zero and nil values",
			params: Params{
				Actual:    nil,
				Canonical: "",
				Limit:     0,
			},
			expect: `actual=<nil>, canonical="", limit=0`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//for checking keys get ordered
			for range 10 {
				if got := tc.params.String(); got != tc.expect {
					t.Fatalf("Params.String() = %q, want %q", got, tc.expect)
				}
			}
		})
	}

}
