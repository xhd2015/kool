package iterm2

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInstall_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runInstall([]string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("help: %v stderr=%q", err, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"--dry-run", "--download-dir", "--download-only", "--via-open", "kool iterm2 install"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q\n%s", want, out)
		}
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "open") && !strings.Contains(low, "gatekeeper") && !strings.Contains(low, "quarantine") {
		t.Errorf("help should mention open/Gatekeeper/quarantine:\n%s", out)
	}
}

func TestRunInstall_ViaOpenIncompatibleWithDownloadOnly(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runInstall([]string{"--via-open", "--download-only", "--download-dir", dir}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for --via-open + --download-only")
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Fatalf("stderr must contain Error:; got %q", stderr.String())
	}
	low := strings.ToLower(stderr.String())
	if !strings.Contains(low, "via-open") || !strings.Contains(low, "download-only") {
		t.Fatalf("stderr should mention both flags; got %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "iTerm2-3_6_11.zip")); !os.IsNotExist(err) {
		t.Fatalf("must not write zip: err=%v", err)
	}
}

func TestRunInstall_DryRunViaOpen(t *testing.T) {
	srv := startFakeITerm2LatestServer(t, []byte("PK\x03\x04fake"))
	withInstallHTTP(t, srv.URL+"/latest", srv.Client())

	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runInstall([]string{"--dry-run", "--via-open", "--download-dir", dir}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dry-run --via-open: %v stderr=%q", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "dry-run:") {
		t.Fatalf("expected dry-run banner:\n%s", out)
	}
	if !strings.Contains(out, "3.6.11") {
		t.Fatalf("expected version:\n%s", out)
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "via-open") && !strings.Contains(low, "open") && !strings.Contains(low, "quarantine") {
		t.Fatalf("dry-run --via-open plan should mention open/via-open/quarantine:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "iTerm2-3_6_11.zip")); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not write zip: err=%v", err)
	}
}

func startFakeITerm2LatestServer(t *testing.T, zipBody []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/latest"):
			http.Redirect(w, r, "http://"+r.Host+"/iTerm2-3_6_11.zip", http.StatusFound)
		case strings.HasSuffix(r.URL.Path, "/iTerm2-3_6_11.zip"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipBody)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func withInstallHTTP(t *testing.T, latestURL string, client *http.Client) {
	t.Helper()
	prevClient := installHTTPClient
	prevURL := installLatestURL
	installHTTPClient = client
	installLatestURL = latestURL
	t.Cleanup(func() {
		installHTTPClient = prevClient
		installLatestURL = prevURL
	})
}

func TestRunInstall_DryRun(t *testing.T) {
	srv := startFakeITerm2LatestServer(t, []byte("PK\x03\x04fake"))
	withInstallHTTP(t, srv.URL+"/latest", srv.Client())

	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runInstall([]string{"--dry-run", "--download-dir", dir}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dry-run: %v stderr=%q", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "dry-run:") {
		t.Fatalf("expected dry-run banner:\n%s", out)
	}
	if !strings.Contains(out, "3.6.11") {
		t.Fatalf("expected version:\n%s", out)
	}
	if !strings.Contains(out, "iTerm2-3_6_11.zip") {
		t.Fatalf("expected zip name under download-dir:\n%s", out)
	}
	if !strings.Contains(out, "--download-dir") {
		t.Fatalf("expected dry-run to note --download-dir:\n%s", out)
	}
	// No zip written on dry-run.
	if _, err := os.Stat(filepath.Join(dir, "iTerm2-3_6_11.zip")); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not write zip: err=%v", err)
	}
}

func TestRunInstall_DownloadOnly(t *testing.T) {
	srv := startFakeITerm2LatestServer(t, []byte("PK\x03\x04not-empty-zip-body"))
	withInstallHTTP(t, srv.URL+"/latest", srv.Client())

	dir := t.TempDir()
	home := t.TempDir()
	prevHome := installHome
	installHome = home
	t.Cleanup(func() { installHome = prevHome })

	zipPath := filepath.Join(dir, "iTerm2-3_6_11.zip")
	var stdout, stderr bytes.Buffer
	err := runInstall([]string{"--download-only", "--download-dir", dir}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("download-only: %v stderr=%q", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "iTerm2  resolve") || !strings.Contains(out, "iTerm2  download") {
		t.Fatalf("expected resolve+download lines:\n%s", out)
	}
	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("zip missing: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("zip empty")
	}
	if _, err := os.Stat(filepath.Join(home, "Applications", "iTerm.app")); !os.IsNotExist(err) {
		t.Fatalf("download-only must not install app: %v", err)
	}
}

func TestRunInstall_UnexpectedArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runInstall([]string{"extra"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "unexpected arguments") {
		t.Fatalf("stderr: %q", stderr.String())
	}
}
