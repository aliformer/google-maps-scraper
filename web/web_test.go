//nolint:testpackage // Need to access unexported server internals
package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, dir string) *Server {
	t.Helper()

	srv, err := New(NewService(nil, dir), ":0")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return srv
}

func TestSPAFallbackAndAPI(t *testing.T) {
	srv := newTestServer(t, t.TempDir())

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantSubstr string
	}{
		{
			name:       "Root returns index.html",
			path:       "/",
			wantStatus: http.StatusOK,
			wantSubstr: "Google Maps Scraper",
		},
		{
			name:       "SPA route returns index.html",
			path:       "/gmaps",
			wantStatus: http.StatusOK,
			wantSubstr: "Google Maps Scraper",
		},
		{
			name:       "API docs endpoint",
			path:       "/api/docs",
			wantStatus: http.StatusOK,
			wantSubstr: "API Documentation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			srv.srv.Handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if !strings.Contains(rec.Body.String(), tt.wantSubstr) {
				t.Errorf("body missing substring %q", tt.wantSubstr)
			}
		})
	}
}

func TestSecurityHeadersAllowMapResources(t *testing.T) {
	handler := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	for _, want := range []string{"tile.openstreetmap.org"} {
		if !strings.Contains(csp, want) {
			t.Fatalf("CSP missing %q: %s", want, csp)
		}
	}
}
