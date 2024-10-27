package bangs

import (
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestQueryUrl_Augment(t *testing.T) {
	t.Parallel()
	type args struct {
		query string
	}
	tests := []struct {
		name    string
		q       QueryURL
		args    args
		want    *url.URL
		wantErr bool
	}{
		{
			name: "Regular Single Word",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "test",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=test",
			},
			wantErr: false,
		},
		{
			name: "Multiple Words",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "test query",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=test+query",
			},
			wantErr: false,
		},
		{
			name: "Special Characters",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "test & query",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=test+%26+query",
			},
			wantErr: false,
		},
		{
			name: "Reserved Characters",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "test?&=+# query",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=test%3F%26%3D%2B%23+query",
			},
			wantErr: false,
		},
		{
			name: "Unicode Characters",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "こんにちは世界",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=%E3%81%93%E3%82%93%E3%81%AB%E3%81%A1%E3%81%AF%E4%B8%96%E7%95%8C",
			},
			wantErr: false,
		},
		{
			name: "Empty Query",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=",
			},
			wantErr: false,
		},
		{
			name: "Special Symbols",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "!@#$%^&*()",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=%21%40%23%24%25%5E%26%2A%28%29",
			},
			wantErr: false,
		},
		{
			name: "Query with Braces",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "test {} query",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=test+%7B%7D+query",
			},
			wantErr: false,
		},
		{
			name: "Bing Search Single Word",
			q:    "https://www.bing.com/search?q={}",
			args: args{
				query: "test",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.bing.com",
				Path:     "/search",
				RawQuery: "q=test",
			},
			wantErr: false,
		},
		{
			name: "DuckDuckGo Search with Multiple Words",
			q:    "https://duckduckgo.com/?q={}",
			args: args{
				query: "test query",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "duckduckgo.com",
				Path:     "/",
				RawQuery: "q=test+query",
			},
			wantErr: false,
		},
		{
			name: "Yahoo Search with Special Characters",
			q:    "https://search.yahoo.com/search?p={}",
			args: args{
				query: "test & query",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "search.yahoo.com",
				Path:     "/search",
				RawQuery: "p=test+%26+query",
			},
			wantErr: false,
		},
		{
			name: "Ecosia Search with Unicode",
			q:    "https://www.ecosia.org/search?q={}",
			args: args{
				query: "こんにちは世界",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.ecosia.org",
				Path:     "/search",
				RawQuery: "q=%E3%81%93%E3%82%93%E3%81%AB%E3%81%A1%E3%81%AF%E4%B8%96%E7%95%8C",
			},
			wantErr: false,
		},
		{
			name: "Empty Query on Bing",
			q:    "https://www.bing.com/search?q={}",
			args: args{
				query: "",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.bing.com",
				Path:     "/search",
				RawQuery: "q=",
			},
			wantErr: false,
		},
		{
			name: "Placeholder in Path",
			q:    "https://www.example.com/{}/search",
			args: args{
				query: "test query",
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "www.example.com",
				Path:   "/test%20query/search",
			},
			wantErr: false,
		},
		{
			name: "Placeholder in Fragment",
			q:    "https://www.example.com/search#{}",
			args: args{
				query: "test",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.example.com",
				Path:     "/search",
				Fragment: "test",
			},
			wantErr: false,
		},
		{
			name: "No Placeholder in QueryUrl",
			q:    "https://www.google.com/search",
			args: args{
				query: "test",
			},
			wantErr: true,
		},
		{
			name: "Multiple Placeholders in QueryUrl",
			q:    "https://www.google.com/search?q={}&hl={}",
			args: args{
				query: "test",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=test&hl=test",
			},
			wantErr: false,
		},
		{
			name: "Placeholder in Host",
			q:    "https://{}.example.com/search",
			args: args{
				query: "test",
			},
			wantErr: true,
		},
		{
			name: "Query with URL-encoded Characters",
			q:    "https://www.google.com/search?q={}",
			args: args{
				query: "test%20query",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "www.google.com",
				Path:     "/search",
				RawQuery: "q=test%2520query",
			},
			wantErr: false,
		},
		{
			name: "Placeholder at the End",
			q:    "https://www.example.com/search/{}",
			args: args{
				query: "test query",
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "www.example.com",
				Path:   "/search/test%20query",
			},
			wantErr: false,
		},
		{
			name: "Placeholder in Scheme",
			q:    "{}://www.google.com/search?q=test",
			args: args{
				query: "https",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.q.Augment(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryUrl.Augment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.String() != tt.want.String() {
					t.Errorf("QueryUrl.Augment() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestQueryUrl_Augment_Errors(t *testing.T) {
	t.Parallel()
	t.Run("URL parsing error", func(t *testing.T) {
		t.Parallel()
		queryURL := QueryURL("://invalid-url")
		_, err := queryURL.Augment("test")
		if err == nil {
			t.Fatalf("expected parse error but got none")
		}
		if _, ok := err.(*url.Error); !ok {
			t.Errorf("expected *url.Error, got %v", err)
		}
	})

	t.Run("No placeholder in URL", func(t *testing.T) {
		t.Parallel()
		queryURL := QueryURL("https://example.com")
		_, err := queryURL.Augment("test")
		if err == nil {
			t.Fatalf("expected AugmentNoPlaceholderError but got none")
		}
		if _, ok := err.(AugmentNoPlaceholderError); !ok {
			t.Errorf("expected AugmentNoPlaceholderError, got %v", err)
		}
	})

	t.Run("Placeholder in unsupported scheme location", func(t *testing.T) {
		t.Parallel()
		queryURL := QueryURL("{}://example.com")
		_, err := queryURL.Augment("test")
		if err == nil {
			t.Fatalf("expected parse error but got none")
		}
		if _, ok := err.(*url.Error); !ok {
			t.Errorf("expected *url.Error due to invalid scheme, got %v", err)
		}
	})

	t.Run("Placeholder in unsupported host location", func(t *testing.T) {
		t.Parallel()
		queryURL := QueryURL("https://{}.example.com")
		_, err := queryURL.Augment("test")
		if err == nil {
			t.Fatalf("expected parse error but got none")
		}
		if _, ok := err.(*url.Error); !ok {
			t.Errorf("expected *url.Error due to invalid host, got %v", err)
		}
	})

	t.Run("Placeholder in unsupported user info location", func(t *testing.T) {
		t.Parallel()
		queryURL := QueryURL("https://user:{}@example.com")
		_, err := queryURL.Augment("test")
		if err == nil {
			t.Fatalf("expected parse error but got none")
		}
		if _, ok := err.(*url.Error); !ok {
			t.Errorf("expected *url.Error due to invalid user info, got %v", err)
		}
	})
}

var runForwardingTests = "false"

func TestAllBangs_ValidForward(t *testing.T) {
	if runForwardingTests != "true" {
		t.Skip("skipping forwarding tests")
	}

	t.Parallel()
	err := Load("../../bangs.yaml")
	if err != nil {
		t.Fatalf("failed to load bangs: %v", err)
	}
	bl := All()

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	for name, entry := range bl.Entries {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			u, err := entry.URL.Augment("test")
			if err != nil {
				t.Errorf("augmenting failed for bang %s: %v", entry.Bang, err)
				return
			}

			req, err := http.NewRequest("GET", u.String(), nil)
			if err != nil {
				t.Errorf("request creation failed for bang %s: %v", entry.Bang, err)
				return
			}
			req.Header.Set("User-Agent", userAgent)

			res, err := client.Do(req)

			if err != nil {
				t.Errorf("request failed for bang %s: %v", entry.Bang, err)
				return
			}
			defer res.Body.Close()

			if res.StatusCode < 200 || (res.StatusCode >= 300 && res.StatusCode != 429 && res.StatusCode != 403) {
				body, _ := io.ReadAll(res.Body) // Read body on error for better output
				t.Errorf("unexpected status code for bang %s and URL %s: got %d, body: %s", u.String(), entry.Bang, res.StatusCode, string(body))
			} else {
				t.Logf("Bang: %s, Status Code: %d", entry.Bang, res.StatusCode)
			}
		})
	}
}

func BenchmarkAugment(b *testing.B) {
	query := "test-query"
	q := QueryURL("https://example.com/search?q={}")

	for i := 0; i < b.N; i++ {
		_, err := q.Augment(query)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
