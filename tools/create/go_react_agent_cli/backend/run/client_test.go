//go:build ignore

package run

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTarget(t *testing.T) {
	cases := []struct {
		name    string
		uri     string
		flags   clientFlags
		want    string
		wantErr bool
	}{
		{name: "bare path uses default port", uri: "/api/counter", want: "http://localhost:8080/api/counter"},
		{name: "query string kept", uri: "/plans?stage=design", want: "http://localhost:8080/plans?stage=design"},
		{name: "port flag overrides", uri: "/x", flags: clientFlags{Port: 9000}, want: "http://localhost:9000/x"},
		{name: "url flag overrides origin", uri: "/x", flags: clientFlags{URL: "http://127.0.0.1:9000/"}, want: "http://127.0.0.1:9000/x"},
		{name: "full url wins", uri: "http://example.com/x?y=1", want: "http://example.com/x?y=1"},
		{name: "https url wins", uri: "https://example.com/x", want: "https://example.com/x"},
		{name: "missing leading slash", uri: "api/counter", wantErr: true},
		{name: "empty", uri: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveTarget(tc.uri, tc.flags)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("resolveTarget(%q) = %q, want %q", tc.uri, got, tc.want)
			}
		})
	}
}

func TestReadPayload(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "body.json")
	if err := os.WriteFile(file, []byte(`{"from":"file"}`), 0644); err != nil {
		t.Fatal(err)
	}

	if got, err := readPayload(`{"from":"literal"}`); err != nil || string(got) != `{"from":"literal"}` {
		t.Fatalf("literal: %q, %v", got, err)
	}
	if got, err := readPayload("@" + file); err != nil || string(got) != `{"from":"file"}` {
		t.Fatalf("@file: %q, %v", got, err)
	}
	if _, err := readPayload("@" + filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("expected error for missing @file")
	}
	if _, err := readPayload("@"); err == nil {
		t.Fatal("expected error for empty @file path")
	}

	// stdin via "-"
	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	if _, err := w.WriteString(`{"from":"stdin"}`); err != nil {
		t.Fatal(err)
	}
	w.Close()
	got, err := readPayload("-")
	os.Stdin = old
	if err != nil || string(got) != `{"from":"stdin"}` {
		t.Fatalf("stdin: %q, %v", got, err)
	}
}

func TestClientVerbsAgainstTestServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			fmt.Fprint(w, "pong")
		case "/api/counter":
			switch r.Method {
			case http.MethodGet:
				fmt.Fprint(w, `{"last":2}`)
			default:
				fmt.Fprintf(w, `{"method":%q}`, r.Method)
			}
		case "/api/plan/1":
			switch r.Method {
			case http.MethodPut:
				fmt.Fprint(w, `{"method":"PUT"}`)
			case http.MethodDelete:
				fmt.Fprint(w, `{"deleted":true}`)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// get prints the body
	out := captureStdout(t, func() {
		if err := doRequest(http.MethodGet, srv.URL+"/ping", nil, requestOptions{}); err != nil {
			t.Errorf("get: %v", err)
		}
	})
	if strings.TrimSpace(out) != "pong" {
		t.Fatalf("get output = %q", out)
	}

	// post without body is silent by default
	out = captureStdout(t, func() {
		if err := doRequest(http.MethodPost, srv.URL+"/api/counter", nil, requestOptions{}); err != nil {
			t.Errorf("post: %v", err)
		}
	})
	if out != "" {
		t.Fatalf("post should be silent without --json, got %q", out)
	}

	// post with --json prints the response
	out = captureStdout(t, func() {
		if err := doRequest(http.MethodPost, srv.URL+"/api/counter", nil, requestOptions{json: true}); err != nil {
			t.Errorf("post --json: %v", err)
		}
	})
	if !strings.Contains(out, `"method":"POST"`) {
		t.Fatalf("post --json output = %q", out)
	}

	// put with a body
	out = captureStdout(t, func() {
		if err := doRequest(http.MethodPut, srv.URL+"/api/plan/1", []byte(`{"title":"x"}`), requestOptions{json: true}); err != nil {
			t.Errorf("put: %v", err)
		}
	})
	if !strings.Contains(out, `"method":"PUT"`) {
		t.Fatalf("put --json output = %q", out)
	}

	// delete
	out = captureStdout(t, func() {
		if err := doRequest(http.MethodDelete, srv.URL+"/api/plan/1", nil, requestOptions{json: true}); err != nil {
			t.Errorf("delete: %v", err)
		}
	})
	if !strings.Contains(out, `"deleted":true`) {
		t.Fatalf("delete --json output = %q", out)
	}
}

func TestDoRequestNon2xxIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"not found: /api/plan/9"}`)
	}))
	defer srv.Close()

	err := doRequest(http.MethodGet, srv.URL+"/api/plan/9", nil, requestOptions{})
	if err == nil {
		t.Fatal("expected error for 404")
	}
	for _, want := range []string{"404", "not found: /api/plan/9"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
}

func TestDoRequestDryRunSendsNothing(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
	}))
	defer srv.Close()

	out := captureStdout(t, func() {
		if err := doRequest(http.MethodPost, srv.URL+"/api/plans", []byte(`{"title":"sketch"}`), requestOptions{dryRun: true}); err != nil {
			t.Errorf("dry-run: %v", err)
		}
	})
	if hits != 0 {
		t.Fatalf("dry-run sent %d requests", hits)
	}
	if !strings.Contains(out, "POST "+srv.URL+"/api/plans") || !strings.Contains(out, `{"title":"sketch"}`) {
		t.Fatalf("dry-run output = %q", out)
	}
}

func TestClientHandlersParseAndDispatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	}))
	defer srv.Close()

	// get with --url routes to the test server
	out := captureStdout(t, func() {
		if err := handleGet([]string{"/ping", "--url", srv.URL}); err != nil {
			t.Errorf("handleGet: %v", err)
		}
	})
	if strings.TrimSpace(out) != "pong" {
		t.Fatalf("handleGet output = %q", out)
	}

	// arg validation
	if err := handleGet([]string{}); err == nil || !strings.Contains(err.Error(), "expected exactly one URI") {
		t.Fatalf("expected URI count error, got %v", err)
	}
	if err := handleGet([]string{"/a", "/b"}); err == nil {
		t.Fatal("expected error for two URIs")
	}
	if err := handleWrite(http.MethodPut, putHelp, []string{"/x", "a", "b"}); err == nil || !strings.Contains(err.Error(), "one URI and one optional JSON body") {
		t.Fatalf("expected body count error, got %v", err)
	}
	if err := handleDelete([]string{"/x", "/y"}); err == nil {
		t.Fatal("expected error for delete with two URIs")
	}

	// --help exits cleanly via ErrHelp
	out = captureStdout(t, func() {
		if err := handleGet([]string{"--help"}); err != nil {
			t.Errorf("get --help: %v", err)
		}
	})
	if !strings.Contains(out, "Usage: __PROJECT_NAME__ get") {
		t.Fatalf("get --help output = %q", out)
	}
}
