package content

import (
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if len(manager.quotes) == 0 {
		t.Error("Quotes should be loaded")
	}

	if len(manager.jokes) == 0 {
		t.Error("Jokes should be loaded")
	}
}

func TestGetRandomQuote(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test getting random quote without category
	quote, err := manager.GetRandom(ContentTypeQuote, "")
	if err != nil {
		t.Fatalf("Failed to get random quote: %v", err)
	}

	if quote == "" {
		t.Error("Quote should not be empty")
	}
}

func TestGetRandomJoke(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test getting random joke without category
	joke, err := manager.GetRandom(ContentTypeJoke, "")
	if err != nil {
		t.Fatalf("Failed to get random joke: %v", err)
	}

	if joke == "" {
		t.Error("Joke should not be empty")
	}
}

func TestGetRandomWithCategory(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name        string
		contentType ContentType
		category    string
		shouldError bool
	}{
		{"Valid quote category", ContentTypeQuote, "inspirational", false},
		{"Valid joke category", ContentTypeJoke, "programming", false},
		{"Invalid quote category", ContentTypeQuote, "nonexistent", true},
		{"Invalid joke category", ContentTypeJoke, "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.GetRandom(tt.contentType, tt.category)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == "" {
					t.Error("Result should not be empty")
				}
			}
		})
	}
}

func TestGetCategories(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test quote categories
	quoteCategories := manager.GetCategories(ContentTypeQuote)
	if len(quoteCategories) == 0 {
		t.Error("Should have quote categories")
	}

	expectedQuoteCategories := []string{"inspirational", "motivational", "life", "success", "wisdom", "love", "happiness", "technology"}
	for _, expected := range expectedQuoteCategories {
		found := false
		for _, category := range quoteCategories {
			if category == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected quote category %s not found", expected)
		}
	}

	// Test joke categories
	jokeCategories := manager.GetCategories(ContentTypeJoke)
	if len(jokeCategories) == 0 {
		t.Error("Should have joke categories")
	}

	expectedJokeCategories := []string{"programming", "science", "dad", "puns", "technology", "work", "animals", "general"}
	for _, expected := range expectedJokeCategories {
		found := false
		for _, category := range jokeCategories {
			if category == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected joke category %s not found", expected)
		}
	}
}

func TestInvalidContentType(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	_, err = manager.GetRandom("invalid", "")
	if err == nil {
		t.Error("Expected error for invalid content type")
	}

	if !strings.Contains(err.Error(), "invalid content type") {
		t.Errorf("Error message should mention invalid content type, got: %v", err)
	}
}
