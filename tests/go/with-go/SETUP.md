# Scenario

**Feature**: kool with-go resolves a pinned GOROOT, lists versions, and execs under it

```
# pin then resolve dest under InstallDir; missing dest uses Install hook
caller goVersion -> ResolveGorootWith -> dest=$InstallDir/<pin> -> existing dir | Install hook

# HandleWith dispatches; ListWith prints go-prefixed versions; ExecGoroot sets child env
args -> HandleWith -> ResolveGorootWith | ListWith | ExecGoroot
goroot + args -> ExecGoroot -> child GOROOT=$abs PATH=$abs/bin:$PATH
```

## Preconditions

- Package `github.com/xhd2015/kool/tools/go/with_go` is the SUT. Injectable
  `HandleWith` / `ResolveGorootWith` / `ListWith` do not exist yet — leaves
  stay compile-RED until they land.
- Parallel-safe: no `os.Chdir` / `t.Chdir` / `os.Setenv` / `t.Setenv` /
  process stdio rewrite. InstallDir is always `t.TempDir()`. Resolve/handle
  leaves inject a recording `Install` hook — never real download network.
- Process cwd is undetermined. Use absolute paths from `t.TempDir()` / `d`.
- `ExecGoroot` writes the child to process stdio (kool-owned). Fake `go`
  scripts write `GOROOT=` / `PATH0=` to `$GOROOT/child.out` so Assert can
  read without rewriting `os.Stdout`.
- Unix script is fine (macOS).

## Steps

1. Grouping/leaf `Setup` sets `req.Op` and fixtures (temp install dir, dest
   dir, fake `$GOROOT/bin/go`, injected `FetchHTML`).
2. Root `Run` dispatches to `ResolveGorootWith`, `HandleWith`, `ListWith`,
   or `ExecGoroot`.

## Context

- Pin table (withgo): `go1.19` / `1.19` → dest `go1.19.13`.
- Fake `go` prints `GOROOT` and first `PATH` entry into `child.out`.
- List stdout is `go` + naked version (`go1.19.13`), one line per HTML div.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func destPin(installDir string) string {
	return filepath.Join(installDir, "go1.19.13")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeFakeGo(t *testing.T, goroot string) {
	t.Helper()
	script := `#!/bin/sh
out="$(cd "$(dirname "$0")/.." && pwd)/child.out"
{
  printf 'GOROOT=%s\n' "$GOROOT"
  IFS=:
  set -- $PATH
  printf 'PATH0=%s\n' "$1"
} > "$out"
`
	writeFile(t, filepath.Join(goroot, "bin", "go"), script)
	if err := os.Chmod(filepath.Join(goroot, "bin", "go"), 0755); err != nil {
		t.Fatal(err)
	}
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
```
