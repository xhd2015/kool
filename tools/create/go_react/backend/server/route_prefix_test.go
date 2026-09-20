//go:build ignore

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
		"////":          "",
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
	handler := MountRoutePrefix("/my-app", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestDevHandlerMountsAPIAndRestoresViteBase(t *testing.T) {
	var gotPath string
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("vite"))
	}))
	defer frontend.Close()

	handler, err := DevHandler(frontend.URL, "/demo")
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/demo/ping", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "pong" {
		t.Fatalf("API response = %d %q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/demo/assets/index.js", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "vite" {
		t.Fatalf("proxy response = %d %q", rec.Code, rec.Body.String())
	}
	if gotPath != "/demo/assets/index.js" {
		t.Fatalf("Vite path = %q, want restored prefix", gotPath)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unprefixed API status = %d, want 404", rec.Code)
	}
}
