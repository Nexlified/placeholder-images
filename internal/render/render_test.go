package render

import "testing"

func TestGetInitials(t *testing.T) {
	cases := []struct {
		name  string
		input string
		exp   string
	}{
		{"empty", "", ""},
		{"single", "alice", "A"},
		{"two words", "alice baker", "AB"},
		{"extra words", "alice baker charlie", "AB"},
		{"mixed spacing", "  alice   baker  ", "AB"},
		{"non letters", "  -alice  123 baker", "-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetInitials(tc.input); got != tc.exp {
				t.Fatalf("expected %q got %q", tc.exp, got)
			}
		})
	}
}
