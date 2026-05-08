package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeRoutePrefix(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"/":             "",
		"my-app":        "/my-app",
		"/my-app":       "/my-app",
		"/my-app/":      "/my-app",
		" /my-app// ":   "/my-app",
		"/nested/app//": "/nested/app",
	}
	for input, want := range cases {
		if got := NormalizeRoutePrefix(input); got != want {
			t.Fatalf("NormalizeRoutePrefix(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestMountRoutePrefixStripsPrefix(t *testing.T) {
	handler := mountRoutePrefix("/my-app", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/my-app/ping", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Body.String(); got != "/ping" {
		t.Fatalf("body = %q, want stripped path", got)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unprefixed status = %d, want 404", rec.Code)
	}
}

func TestPrepareFrontendHTML(t *testing.T) {
	html := []byte(`<html><head><script type="module">import "/assets/index.js"</script><link href="/assets/index.css"></head><body></body></html>`)
	out := string(prepareFrontendHTML(html, "my-app", true))
	for _, want := range []string{
		`window.__KOOL_ROUTE_PREFIX__="/my-app"`,
		`import "/my-app/assets/index.js"`,
		`href="/my-app/assets/index.css"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("prepared HTML missing %q: %s", want, out)
		}
	}
}
