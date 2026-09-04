package modcache

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestHandleSeedHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := HandleWith([]string{"seed", "--help"}, HandleOpts{Stdout: &stdout, Stderr: &stderr}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "kool go modcache seed") {
		t.Fatalf("help:\n%s", stdout.String())
	}
}

func TestHandleSeedMissingGoMod(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := HandleWith([]string{"seed", dir}, HandleOpts{Stdout: &stdout, Stderr: &stderr})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "go.mod") {
		t.Fatalf("stderr: %s", stderr.String())
	}
}

func TestHandleSeedLatest(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	scene := t.TempDir()
	modCache := filepath.Join(scene, "gomodcache")
	t.Cleanup(func() { chmodTreeWritable(modCache) })
	lib := filepath.Join(scene, "lib")
	mustMkdir(t, lib)
	mustWrite(t, filepath.Join(lib, "go.mod"), "module github.com/xhd2015/wrk-seed-expt\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(lib, "hello.go"), "package seedexpt\n")
	git(t, lib, "init", "-b", "main")
	git(t, lib, "config", "user.name", "expt")
	git(t, lib, "config", "user.email", "expt@example.com")
	git(t, lib, "add", ".")
	git(t, lib, "commit", "-m", "init")
	git(t, lib, "tag", "v0.0.1")

	t.Setenv("GOMODCACHE", modCache)
	t.Setenv("GOCACHE", filepath.Join(scene, "gocache"))
	t.Setenv("GOPATH", filepath.Join(scene, "gopath"))
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("https_proxy", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")

	var stdout, stderr bytes.Buffer
	if err := HandleWith([]string{"seed", lib}, HandleOpts{Stdout: &stdout, Stderr: &stderr}); err != nil {
		t.Fatalf("%v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
	}
	spine := regexp.MustCompile(`(?m)^\[[0-9]+/`)
	matches := spine.FindAllStringIndex(stderr.String(), -1)
	if len(matches) != 2 {
		t.Fatalf("want 2 spine lines, got %d\n%s", len(matches), stderr.String())
	}
	if strings.Count(stderr.String(), "[2/2]") != 1 {
		t.Fatalf("duplicate [2/2]:\n%s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "seeded github.com/xhd2015/wrk-seed-expt@v0.0.1") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	// Blank line before product: stderr should end with blank before we check —
	// product is on stdout; ensure stderr has trailing blank after last stage.
	if !strings.HasSuffix(stderr.String(), "\n\n") && !strings.Contains(stderr.String(), "\n\n") {
		t.Fatalf("expected blank line on stderr before product:\n%q", stderr.String())
	}
}

func TestHandleSeedDryRun(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	scene := t.TempDir()
	lib := filepath.Join(scene, "lib")
	mustMkdir(t, lib)
	mustWrite(t, filepath.Join(lib, "go.mod"), "module github.com/xhd2015/wrk-seed-expt\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(lib, "hello.go"), "package seedexpt\n")
	git(t, lib, "init", "-b", "main")
	git(t, lib, "config", "user.name", "expt")
	git(t, lib, "config", "user.email", "expt@example.com")
	git(t, lib, "add", ".")
	git(t, lib, "commit", "-m", "init")
	git(t, lib, "tag", "v0.0.1")

	var stdout, stderr bytes.Buffer
	if err := HandleWith([]string{"seed", "--dry-run", lib}, HandleOpts{Stdout: &stdout, Stderr: &stderr}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stderr.String(), "skip (dry-run)") {
		t.Fatalf("banned skip (dry-run):\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "would: go mod download") {
		t.Fatalf("stderr missing would:\n%s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "would: seed github.com/xhd2015/wrk-seed-expt@v0.0.1") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestHandleSeedHeadUntagged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	scene := t.TempDir()
	lib := filepath.Join(scene, "lib")
	mustMkdir(t, lib)
	mustWrite(t, filepath.Join(lib, "go.mod"), "module github.com/xhd2015/wrk-seed-expt\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(lib, "hello.go"), "package seedexpt\n")
	git(t, lib, "init", "-b", "main")
	git(t, lib, "config", "user.name", "expt")
	git(t, lib, "config", "user.email", "expt@example.com")
	git(t, lib, "add", ".")
	git(t, lib, "commit", "-m", "init")
	git(t, lib, "tag", "v0.0.1")
	mustWrite(t, filepath.Join(lib, "hello.go"), "package seedexpt\n\nfunc X() {}\n")
	git(t, lib, "add", ".")
	git(t, lib, "commit", "-m", "untagged")

	var stdout, stderr bytes.Buffer
	err := HandleWith([]string{"seed", "--tag", "HEAD", lib}, HandleOpts{Stdout: &stdout, Stderr: &stderr})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "HEAD is not tagged") {
		t.Fatalf("stderr=%s", stderr.String())
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func chmodTreeWritable(root string) {
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		mode := info.Mode()
		if mode.IsDir() {
			_ = os.Chmod(path, mode|0o700)
		} else {
			_ = os.Chmod(path, mode|0o600)
		}
		return nil
	})
}
