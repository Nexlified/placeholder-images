package utils

import "testing"

func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		name  string
		input string
		def   int
		exp   int
	}{
		{"empty", "", 10, 10},
		{"valid", "42", 10, 42},
		{"zero", "0", 10, 10},
		{"negative", "-5", 10, 10},
		{"invalid", "abc", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseIntOrDefault(tt.input, tt.def); got != tt.exp {
				t.Fatalf("expected %d, got %d", tt.exp, got)
			}
		})
	}
}

func TestGenerateColorHashDeterministic(t *testing.T) {
	seed := "Jane Doe"
	first := GenerateColorHash(seed)
	second := GenerateColorHash(seed)
	if first != second {
		t.Fatalf("expected deterministic hash, got %s and %s", first, second)
	}
}
