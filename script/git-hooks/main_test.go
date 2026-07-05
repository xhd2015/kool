package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPreCommitAutoInstallFrontendDeps(t *testing.T) {
	dir := t.TempDir()

	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.email", "test@example.com")
	runCmd(t, dir, "git", "config", "user.name", "Test")

	writeFile(t, filepath.Join(dir, "go.mod"), "module test\ngo 1.23\n")
	writeFile(t, filepath.Join(dir, "main.go"), "package main\nfunc main() {}\n")

	frontendDir := filepath.Join(dir, "tools", "web", "react")
	if err := os.MkdirAll(frontendDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(frontendDir, "package.json"), `{"scripts":{"build":"echo ok"}}`)

	runCmd(t, frontendDir, "pnpm", "install")
	if _, err := os.Stat(filepath.Join(frontendDir, "node_modules")); os.IsNotExist(err) {
		t.Fatal("pnpm install did not create node_modules")
	}

	if err := os.RemoveAll(filepath.Join(frontendDir, "node_modules")); err != nil {
		t.Fatal(err)
	}

	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "initial")

	writeFile(t, filepath.Join(dir, "main.go"), "package main\nfunc main() { println(1) }\n")
	runCmd(t, dir, "git", "add", "main.go")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	err = preCommitCheck(false, false)
	if err != nil {
		t.Fatalf("preCommitCheck failed when node_modules missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(frontendDir, "node_modules")); os.IsNotExist(err) {
		t.Fatal("expected node_modules to be created by auto-install")
	}

	writeFile(t, filepath.Join(dir, "main.go"), "package main\nfunc main() { println(2) }\n")
	runCmd(t, dir, "git", "add", "main.go")

	err = preCommitCheck(false, false)
	if err != nil {
		t.Fatalf("preCommitCheck failed when node_modules exists: %v", err)
	}
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
