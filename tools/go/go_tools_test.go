package go_tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstGOPATHBinDirUsesFirstPath(t *testing.T) {
	sep := string(os.PathListSeparator)
	first := filepath.Join("tmp", "first")
	second := filepath.Join("tmp", "second")

	binDir, err := firstGOPATHBinDir(first + sep + second)
	if err != nil {
		t.Fatalf("firstGOPATHBinDir returned error: %v", err)
	}

	want := filepath.Join(first, "bin")
	if binDir != want {
		t.Fatalf("bin dir = %q, want %q", binDir, want)
	}
}

func TestFirstGOPATHBinDirSkipsEmptyPathEntries(t *testing.T) {
	sep := string(os.PathListSeparator)
	first := filepath.Join("tmp", "first")

	binDir, err := firstGOPATHBinDir(sep + first)
	if err != nil {
		t.Fatalf("firstGOPATHBinDir returned error: %v", err)
	}

	want := filepath.Join(first, "bin")
	if binDir != want {
		t.Fatalf("bin dir = %q, want %q", binDir, want)
	}
}

func TestResolveRebuildOutputPathFallsBackToGOPATH(t *testing.T) {
	firstGOPATH := t.TempDir()
	secondGOPATH := t.TempDir()
	t.Setenv("PATH", t.TempDir())
	t.Setenv("GOPATH", firstGOPATH+string(os.PathListSeparator)+secondGOPATH)

	binaryName := "kool-rebuild-test-definitely-missing"
	outputPath, err := resolveRebuildOutputPath(binaryName, false)
	if err != nil {
		t.Fatalf("resolveRebuildOutputPath returned error: %v", err)
	}

	want, err := filepath.Abs(filepath.Join(firstGOPATH, "bin", executableName(binaryName)))
	if err != nil {
		t.Fatal(err)
	}
	if outputPath != want {
		t.Fatalf("output path = %q, want %q", outputPath, want)
	}
	if _, err := os.Stat(filepath.Join(firstGOPATH, "bin")); err != nil {
		t.Fatalf("first GOPATH bin dir was not created: %v", err)
	}
}

func TestResolveRebuildOutputPathForceGOPATHRequiresGOPATH(t *testing.T) {
	t.Setenv("GOPATH", "")

	_, err := resolveRebuildOutputPath("kool-rebuild-test", true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "GOPATH is not set") {
		t.Fatalf("error = %q, want GOPATH message", err)
	}
}
