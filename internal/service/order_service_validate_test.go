package service

import "testing"

func TestNormalizeGuestPhone_CanonicalizesFormatting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain digits", input: "13800138000", want: "13800138000"},
		{name: "spaces and dashes", input: " 138-0013 8000 ", want: "13800138000"},
		{name: "country code and parentheses", input: "(+86) 138-0013-8000", want: "+8613800138000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeGuestPhone(tc.input)
			if err != nil {
				t.Fatalf("normalizeGuestPhone returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalizeGuestPhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
