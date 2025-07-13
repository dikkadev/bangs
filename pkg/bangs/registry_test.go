package bangs

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var sizes = []int{10, 20, 100, 500, 1e3, 1e4, 5e4, 1e5, 5e5, 1e6, 5e6, 1e7, 5e7, 1e8}

func generateRandomBangs(N int) BangList {
	bl := BangList{
		Entries: make(map[string]Entry, N),
		byBang:  make(map[string]Entry, N),
		len:     N,
	}

	alphabet := "abcdefghijklmnopqrstuvwxyz"
	totalGenerated := 0
	length := 1

	for totalGenerated < N {
		possibleCombos := int(math.Pow(26, float64(length)))
		totalGenerated += possibleCombos
		length++
	}

	length--

	generatedCount := 0
	for i := 0; i < length; i++ {
		for j := 0; j < 26; j++ {
			if generatedCount >= N {
				break
			}
			bang := string(alphabet[j])
			for k := 0; k < i; k++ {
				bang += string(alphabet[j])
			}
			bl.Entries[bang] = Entry{
				Bang:     bang,
				URL:      QueryURL("https://www.google.com/search?q={}"),
				Category: "",
			}
			bl.byBang[bang] = bl.Entries[bang]
			generatedCount++
		}
	}

	return bl
}

func TestRegistry_DefaultForward(t *testing.T) {
	// Create test registry with some bangs
	registry := &Registry{
		Entries: BangList{
			Entries: map[string]Entry{
				"Google": {
					Bang: "g",
					URL:  "https://www.google.com/search?q={}",
				},
				"GitHub": {
					Bang: "gh",
					URL:  "https://github.com/search?q={}",
				},
				"StackOverflow": {
					Bang: "so",
					URL:  "https://stackoverflow.com/search?q={}",
				},
			},
			byBang: map[string]Entry{
				"g": {
					Bang: "g",
					URL:  "https://www.google.com/search?q={}",
				},
				"gh": {
					Bang: "gh",
					URL:  "https://github.com/search?q={}",
				},
				"so": {
					Bang: "so",
					URL:  "https://stackoverflow.com/search?q={}",
				},
			},
		},
	}

	tests := []struct {
		name           string
		defaultValue   string
		query          string
		expectedStatus int
		expectedURL    string
		expectMultiple bool
	}{
		{
			name:           "Traditional URL default",
			defaultValue:   "https://www.google.com/search?q={}",
			query:          "test query",
			expectedStatus: http.StatusFound,
			expectedURL:    "https://www.google.com/search?q=test+query",
			expectMultiple: false,
		},
		{
			name:           "Single bang reference",
			defaultValue:   "g",
			query:          "test query",
			expectedStatus: http.StatusFound,
			expectedURL:    "https://www.google.com/search?q=test+query",
			expectMultiple: false,
		},
		{
			name:           "Multi-bang reference",
			defaultValue:   "g+gh",
			query:          "test query",
			expectedStatus: http.StatusOK,
			expectedURL:    "",
			expectMultiple: true,
		},
		{
			name:           "Multi-bang with spaces",
			defaultValue:   "g + gh + so",
			query:          "test query",
			expectedStatus: http.StatusOK,
			expectedURL:    "",
			expectMultiple: true,
		},
		{
			name:           "Invalid bang reference",
			defaultValue:   "invalid",
			query:          "test query",
			expectedStatus: http.StatusInternalServerError,
			expectedURL:    "",
			expectMultiple: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry.Default = QueryURL(tt.defaultValue)

			req := httptest.NewRequest("GET", "/bang", nil)
			w := httptest.NewRecorder()

			err := registry.DefaultForward(tt.query, w, req)

			if tt.expectedStatus == http.StatusInternalServerError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectMultiple {
				// For multi-bang, check that we get HTML with JavaScript
				body := w.Body.String()
				if !strings.Contains(body, "window.open") {
					t.Errorf("Expected multi-bang HTML with window.open, got: %s", body)
				}
				if !strings.Contains(body, "window.location.href") {
					t.Errorf("Expected multi-bang HTML with window.location.href, got: %s", body)
				}
			} else if tt.expectedURL != "" {
				// For single redirects, check the Location header
				location := w.Header().Get("Location")
				if location != tt.expectedURL {
					t.Errorf("Expected redirect to %s, got %s", tt.expectedURL, location)
				}
			}
		})
	}
}

func BenchmarkPrepareInputPreComp(b *testing.B) {
	for _, size := range sizes {
		bl := generateRandomBangs(size)

		allBangs := make([]string, 0, len(bl.Entries))
		for _, bang := range bl.Entries {
			allBangs = append(allBangs, bang.Bang)
		}

		//all inputs
		inputs := make([]string, len(allBangs))
		for i, bang := range allBangs {
			inputs[i] = fmt.Sprintf("!%s some query text", bang)
		}

		b.ResetTimer()
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

		b.Run(fmt.Sprintf("PrepareInput-%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, input := range inputs {
					// _, _, err := bl.PrepareInputNaive(input)
					_, _, err := bl.PrepareInput(input)
					if err != nil {
						b.Errorf("PrepareInputPreComp failed: %v", err)
					}
				}
			}
		})
	}
}

func BenchmarkPrepareInputNaive(b *testing.B) {
	for _, size := range sizes {
		bl := generateRandomBangs(size)

		allBangs := make([]string, 0, len(bl.Entries))
		for _, bang := range bl.Entries {
			allBangs = append(allBangs, bang.Bang)
		}

		inputs := make([]string, len(allBangs))
		for i, bang := range allBangs {
			inputs[i] = fmt.Sprintf("!%s some query text", bang)
		}

		b.ResetTimer()
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

		b.Run(fmt.Sprintf("PrepareInput-%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, input := range inputs {
					_, _, err := bl.PrepareInputNaive(input)
					// _, _, err := bl.PrepareInput(input)
					if err != nil {
						b.Errorf("PrepareInputNaive failed: %v", err)
					}
				}
			}
		})
	}
}
