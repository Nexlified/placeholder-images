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

func TestGetContrastColorWithGradient(t *testing.T) {
	cases := []struct {
		name  string
		input string
		exp   string
	}{
		{"light gradient", "ffffff,cccccc", "000000"},
		{"dark gradient", "000000,333333", "ffffff"},
		{"red to blue", "ff0000,0000ff", "ffffff"},
		{"single color still works", "ffffff", "000000"},
		{"single dark color", "000000", "ffffff"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetContrastColor(tc.input); got != tc.exp {
				t.Fatalf("expected %q got %q", tc.exp, got)
			}
		})
	}
}

func TestDrawImageWithGradient(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	// Test that gradient image generation doesn't error
	_, err = r.DrawImageWithFormat(400, 300, "ff0000,0000ff", "ffffff", "Test", false, false, FormatPNG)
	if err != nil {
		t.Fatalf("failed to draw image with gradient: %v", err)
	}

	// Test with single color (existing behavior)
	_, err = r.DrawImageWithFormat(400, 300, "ff0000", "ffffff", "Test", false, false, FormatPNG)
	if err != nil {
		t.Fatalf("failed to draw image with solid color: %v", err)
	}
}
