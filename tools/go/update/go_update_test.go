package update

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/kool/tools/go/resolve"
)

func TestStripSubDirFromTag(t *testing.T) {
	testCases := []struct {
		name       string
		tag        string
		modulePath string
		expected   string
	}{
		{
			name:       "sub-directory tag with matching module path",
			tag:        "sub/module-a/v1.20.1",
			modulePath: "github.com/example/repo/sub/module-a",
			expected:   "v1.20.1",
		},
		{
			name:       "sub-directory tag with matching module path (exact match)",
			tag:        "module-a/v2.0.0",
			modulePath: "github.com/example/repo/module-a",
			expected:   "v2.0.0",
		},
		{
			name:       "regular tag without sub-directory",
			tag:        "v1.20.1",
			modulePath: "github.com/example/repo",
			expected:   "v1.20.1",
		},
		{
			name:       "sub-directory tag with non-matching module path",
			tag:        "sub/module-a/v1.20.1",
			modulePath: "github.com/example/repo/other/module",
			expected:   "sub/module-a/v1.20.1",
		},
		{
			name:       "empty tag",
			tag:        "",
			modulePath: "github.com/example/repo",
			expected:   "",
		},
		{
			name:       "deeper sub-directory",
			tag:        "path/to/module/v2.0.0",
			modulePath: "github.com/example/repo/path/to/module",
			expected:   "v2.0.0",
		},
		{
			name:       "tag with single part (no slash)",
			tag:        "v1.0.0",
			modulePath: "github.com/example/repo/some/module",
			expected:   "v1.0.0",
		},
		{
			name:       "partial match should not strip",
			tag:        "sub/module-a/v1.20.1",
			modulePath: "github.com/example/repo/sub/module-b",
			expected:   "sub/module-a/v1.20.1",
		},
		{
			name:       "module path without sub-directory",
			tag:        "sub/module/v1.0.0",
			modulePath: "github.com/example/repo",
			expected:   "sub/module/v1.0.0",
		},
		{
			name:       "complex nested sub-directory",
			tag:        "internal/pkg/utils/v3.1.4",
			modulePath: "github.com/company/project/internal/pkg/utils",
			expected:   "v3.1.4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripSubDirFromTag(tc.tag, tc.modulePath)
			if result != tc.expected {
				t.Errorf("stripSubDirFromTag(%q, %q) = %q, expected %q",
					tc.tag, tc.modulePath, result, tc.expected)
			}
		})
	}
}

func TestUpdateUsesLatestRootTagInsteadOfHeadPseudoVersion(t *testing.T) {
	workspace := t.TempDir()
	target := filepath.Join(workspace, "kode-ai")
	consumer := filepath.Join(workspace, "spl")
	initKodeAIRepo(t, target)
	initConsumerModule(t, consumer, "github.com/xhd2015/kode-ai", "v0.0.1", target)

	withDir(t, consumer, func() {
		if err := Update(target); err != nil {
			t.Fatal(err)
		}
	})

	modInfo, err := resolve.GetModuleInfo(consumer)
	if err != nil {
		t.Fatal(err)
	}
	if got := requireVersion(modInfo, "github.com/xhd2015/kode-ai"); got != "v0.0.44" {
		t.Fatalf("root module version = %q, want v0.0.44", got)
	}
	if hasReplace(modInfo, "github.com/xhd2015/kode-ai") {
		t.Fatalf("root module replace was not dropped")
	}
}

func TestUpdateUsesLatestSubmoduleTagInsteadOfHeadPseudoVersion(t *testing.T) {
	workspace := t.TempDir()
	target := filepath.Join(workspace, "kode-ai")
	typesDir := filepath.Join(target, "types")
	consumer := filepath.Join(workspace, "spl")
	initKodeAIRepo(t, target)
	initConsumerModule(t, consumer, "github.com/xhd2015/kode-ai/types", "v0.0.1", typesDir)

	withDir(t, consumer, func() {
		if err := Update(typesDir); err != nil {
			t.Fatal(err)
		}
	})

	modInfo, err := resolve.GetModuleInfo(consumer)
	if err != nil {
		t.Fatal(err)
	}
	if got := requireVersion(modInfo, "github.com/xhd2015/kode-ai/types"); got != "v0.0.13" {
		t.Fatalf("types module version = %q, want v0.0.13", got)
	}
	if hasReplace(modInfo, "github.com/xhd2015/kode-ai/types") {
		t.Fatalf("types module replace was not dropped")
	}
}

func initKodeAIRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "types"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "go.mod"), "module github.com/xhd2015/kode-ai\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dir, "kode.go"), "package kode\n")
	writeFile(t, filepath.Join(dir, "types", "go.mod"), "module github.com/xhd2015/kode-ai/types\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dir, "types", "types.go"), "package types\n")
	mustRun(t, dir, "git", "init")
	mustRun(t, dir, "git", "config", "user.email", "test@example.com")
	mustRun(t, dir, "git", "config", "user.name", "Test User")
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "release")
	mustRun(t, dir, "git", "tag", "v0.0.44")
	mustRun(t, dir, "git", "tag", "types/v0.0.13")
	writeFile(t, filepath.Join(dir, "README.md"), "update README\n")
	mustRun(t, dir, "git", "add", "README.md")
	mustRun(t, dir, "git", "commit", "-m", "update README")
}

func initConsumerModule(t *testing.T, dir string, modulePath string, version string, replaceDir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/spl\n\ngo 1.22\n")
	mustRun(t, dir, "go", "mod", "edit", "-require="+modulePath+"@"+version)
	mustRun(t, dir, "go", "mod", "edit", "-replace="+modulePath+"="+replaceDir)
}

func requireVersion(modInfo *resolve.ModuleInfo, modulePath string) string {
	for _, req := range modInfo.Require {
		if req.Path == modulePath {
			return req.Version
		}
	}
	return ""
}

func hasReplace(modInfo *resolve.ModuleInfo, modulePath string) bool {
	for _, repl := range modInfo.Replace {
		if repl.Old.Path == modulePath {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}

func withDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(old); err != nil {
			t.Fatal(err)
		}
	}()
	fn()
}
