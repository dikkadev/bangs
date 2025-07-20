package main

import (
	"bangs/pkg/bangs"
	"os"
	"testing"
)

func TestFindBangName(t *testing.T) {
	// Create a temporary bangs.yaml for testing
	tempFile, err := os.CreateTemp("", "bangs_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testConfig := `
GitHub:
  bang: 'gh'
  url: 'https://github.com/search?q={}'
  description: 'Search GitHub repositories'
  category: 'Development'

Google:
  bang: 'g'
  url: 'https://www.google.com/search?q={}'
  description: 'Search Google'
  category: 'Search'
`

	_, err = tempFile.WriteString(testConfig)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tempFile.Close()

	// Load the test configuration
	err = bangs.Load(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load test bangs: %v", err)
	}

	tests := []struct {
		bangChars string
		expected  string
	}{
		{"gh", "GitHub"},
		{"g", "Google"},
		{"nonexistent", "nonexistent"}, // Should fallback to bang chars
	}

	for _, test := range tests {
		result := findBangName(test.bangChars)
		if result != test.expected {
			t.Errorf("findBangName(%q) = %q, expected %q", test.bangChars, result, test.expected)
		}
	}
}

func TestBangResultStruct(t *testing.T) {
	result := BangResult{
		Bang: "gh",
		Name: "GitHub",
		URL:  "https://github.com/search?q=test",
	}

	if result.Bang != "gh" {
		t.Errorf("Expected Bang to be 'gh', got %q", result.Bang)
	}
	if result.Name != "GitHub" {
		t.Errorf("Expected Name to be 'GitHub', got %q", result.Name)
	}
	if result.URL != "https://github.com/search?q=test" {
		t.Errorf("Expected URL to be 'https://github.com/search?q=test', got %q", result.URL)
	}
}

func TestMultiBangResultStruct(t *testing.T) {
	result := MultiBangResult{
		Results: []BangResult{
			{Bang: "g", Name: "Google", URL: "https://google.com/search?q=test"},
			{Bang: "gh", Name: "GitHub", URL: "https://github.com/search?q=test"},
		},
		Errors: []string{"Bang 'invalid' not found"},
	}

	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0] != "Bang 'invalid' not found" {
		t.Errorf("Expected error message 'Bang 'invalid' not found', got %q", result.Errors[0])
	}
}

func TestBangInfoStruct(t *testing.T) {
	info := BangInfo{
		Name:        "GitHub",
		Bang:        "gh",
		Description: "Search GitHub repositories",
		Category:    "Development",
		URL:         "https://github.com/search?q={}",
	}

	if info.Name != "GitHub" {
		t.Errorf("Expected Name to be 'GitHub', got %q", info.Name)
	}
	if info.Bang != "gh" {
		t.Errorf("Expected Bang to be 'gh', got %q", info.Bang)
	}
	if info.Category != "Development" {
		t.Errorf("Expected Category to be 'Development', got %q", info.Category)
	}
}
