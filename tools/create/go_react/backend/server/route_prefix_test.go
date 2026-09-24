//go:build ignore

package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeRoutePrefix(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"/":             "",
		"////":          "",
		"demo":          "/demo",
		"/demo":         "/demo",
		"/demo/":        "/demo",
		" /demo// ":     "/demo",
		"/nested/app//": "/nested/app",
	}
	for input, want := range cases {
		if got := NormalizeRoutePrefix(input); got != want {
			t.Errorf("NormalizeRoutePrefix(%q) = %q, want %q", input, got, want)
		}
	}
	if got := NormalizeRoute(RouteOptions{Prefix: " /demo// "}); got.Prefix != "/demo" || got.KeepRoot {
		t.Errorf("NormalizeRoute = %+v", got)
	}
}

func TestValidateRoutePrefix(t *testing.T) {
	for _, ok := range []string{"", "/", "demo", "/demo", "/a-b/c_d.e"} {
		if err := ValidateRoutePrefix(ok); err != nil {
			t.Errorf("ValidateRoutePrefix(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"a b", "/a b", "/a?b", "/a#b", "/a%2Fb"} {
		if err := ValidateRoutePrefix(bad); err == nil {
			t.Errorf("ValidateRoutePrefix(%q) = nil, want an error", bad)
		}
	}
}

func TestMountRoutePrefixStripsAndInjectsPrefix(t *testing.T) {
	handler := MountRoutePrefix(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(RequestRoutePrefix(r) + " " + r.URL.Path))
	}), RouteOptions{Prefix: "/demo"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/demo/api/items", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Body.String(); got != "/demo /api/items" {
		t.Fatalf("body = %q", got)
	}

	// The bare prefix redirects into it, so relative links keep working.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/demo", nil))
	if rec.Code != http.StatusTemporaryRedirect || rec.Header().Get("Location") != "/demo/" {
		t.Fatalf("bare prefix = %d %q", rec.Code, rec.Header().Get("Location"))
	}

	// Off-prefix 404s unless the root route is kept.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/items", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unprefixed status = %d, want 404", rec.Code)
	}
}

func TestMountRoutePrefixKeepRootServesBothRoutes(t *testing.T) {
	handler := MountRoutePrefix(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(RequestRoutePrefix(r) + " " + r.URL.Path))
	}), RouteOptions{Prefix: "/demo", KeepRoot: true})

	for _, tc := range []struct{ path, want string }{
		{"/demo/", "/demo /"},
		{"/demo/api/items", "/demo /api/items"},
		{"/", " /"},
		{"/api/items", " /api/items"},
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", tc.path, rec.Code)
		}
		if got := rec.Body.String(); got != tc.want {
			t.Errorf("%s body = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestMountRoutePrefixEmptyIsIdentity(t *testing.T) {
	handler := MountRoutePrefix(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(RequestRoutePrefix(r) + " " + r.URL.Path))
	}), RouteOptions{})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/items", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != " /api/items" {
		t.Fatalf("root mount = %d %q", rec.Code, rec.Body.String())
	}
}

func TestPrepareFrontendHTMLPrefixesRootAssetsOnce(t *testing.T) {
	html := []byte(`<html><head>` +
		`<script type="module">import "/assets/index.js"</script>` +
		`<link rel="icon" href="/vite.svg">` +
		`<script type="module" crossorigin src="/assets/index-abc.js"></script>` +
		`<link rel="stylesheet" crossorigin href="/assets/index-abc.css">` +
		`</head><body></body></html>`)

	prefixed := string(prepareFrontendHTML(html, "demo", true))
	for _, want := range []string{
		`window.__KOOL_ROUTE_PREFIX__="/demo"`,
		`import "/demo/assets/index.js"`,
		`href="/demo/vite.svg"`,
		`src="/demo/assets/index-abc.js"`,
		`href="/demo/assets/index-abc.css"`,
	} {
		if !strings.Contains(prefixed, want) {
			t.Errorf("prefixed HTML missing %q:\n%s", want, prefixed)
		}
	}

	// A dev server started with --base already prefixed its own HTML; the
	// rewrite must not double it.
	twice := string(prepareFrontendHTML([]byte(prefixed), "/demo", true))
	if strings.Contains(twice, "/demo/demo/") {
		t.Fatalf("double prefix:\n%s", twice)
	}

	// The root route keeps asset URLs untouched and publishes an empty prefix.
	root := string(prepareFrontendHTML(html, "", true))
	for _, want := range []string{`window.__KOOL_ROUTE_PREFIX__=""`, `src="/assets/index-abc.js"`} {
		if !strings.Contains(root, want) {
			t.Errorf("root HTML missing %q:\n%s", want, root)
		}
	}
}

func TestProxyDevRestoresPrefixForViteAndInjectsHTML(t *testing.T) {
	var vitePath string
	vite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vitePath = r.URL.Path
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><script type="module" src="/demo/src/main.tsx"></script></head><body></body></html>`))
	}))
	defer vite.Close()

	target, err := url.Parse(vite.URL)
	if err != nil {
		t.Fatal(err)
	}
	route := RouteOptions{Prefix: "/demo", KeepRoot: true}
	mux := http.NewServeMux()
	if err := proxyDevTarget(mux, target, route); err != nil {
		t.Fatal(err)
	}
	handler := MountRoutePrefix(mux, route)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/demo/app", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if vitePath != "/demo/app" {
		t.Errorf("vite path = %q, want the prefix restored", vitePath)
	}
	if body := rec.Body.String(); !strings.Contains(body, `window.__KOOL_ROUTE_PREFIX__="/demo"`) {
		t.Errorf("prefixed HTML missing the global:\n%s", body)
	}

	// The same process serves the direct domain: Vite still gets its base path,
	// while the page learns it is on the root route.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("root status = %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `window.__KOOL_ROUTE_PREFIX__=""`) {
		t.Errorf("root HTML should clear the prefix:\n%s", body)
	}
}
