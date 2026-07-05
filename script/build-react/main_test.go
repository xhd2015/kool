package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

func TestResolveUsesPackageManagerField(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{
  "name": "demo",
  "packageManager": "pnpm@11.10.0"
}`)
	writeFile(t, filepath.Join(dir, "bun.lock"), "{}")

	got, err := npm.Resolve(dir, "auto")
	if err != nil {
		t.Fatalf("npm.Resolve() error: %v", err)
	}
	if got != npm.ManagerPnpm {
		t.Fatalf("npm.Resolve() = %q, want pnpm", got)
	}
}

func TestResolveUsesPnpmLockfile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{"name":"demo"}`)
	writeFile(t, filepath.Join(dir, "pnpm-lock.yaml"), "lockfileVersion: '9.0'\n")

	got, err := npm.Resolve(dir, "auto")
	if err != nil {
		t.Fatalf("npm.Resolve() error: %v", err)
	}
	if got != npm.ManagerPnpm {
		t.Fatalf("npm.Resolve() = %q, want pnpm", got)
	}
}

func TestResolveExplicitPnpm(t *testing.T) {
	dir := t.TempDir()
	got, err := npm.Resolve(dir, "pnpm")
	if err != nil {
		t.Fatalf("npm.Resolve() error: %v", err)
	}
	if got != npm.ManagerPnpm {
		t.Fatalf("npm.Resolve() = %q, want pnpm", got)
	}
}

func TestResolveRejectsUnknownManager(t *testing.T) {
	dir := t.TempDir()
	if _, err := npm.Resolve(dir, "yarnberry"); err == nil {
		t.Fatal("npm.Resolve() expected error for unknown manager")
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