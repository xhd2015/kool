# Scenario

**Feature**: kool sandbox build + sealed-binary run (P1 package, P2 unpack/exec, home-linked, runtime-load-devbox seal + load-devbox runtime)

```
# help
user -> kool sandbox [--help | build --help]
  -> usage on stdout, exit 0 (build help lists --home-linked, --runtime-load-devbox)

# validation
user -> kool sandbox build [missing -o | empty pack | bad --env | missing --file source | relative --runtime-load-devbox]
  -> stderr error, non-zero; no sealed artifact required

# build happy
user -> kool sandbox build -o OUT [-i DIR] [--file L=R]... [--env K=V]... [--home-linked] [--runtime-load-devbox ABS]...
  -> one-time RSA + AES-GCM sealed PackBlob (optional home_linked / runtime_load_devbox) embedded in OUT

# inspect (optional post-build)
user -> kool sandbox inspect OUT
  -> paths, content hashes, env keys (never secret values); runtime-load-devbox paths when sealed

# sealed run (P2)
user -> KOOL_SANDBOX_ROOT=PARENT [HOME=fake-home] ./OUT [--load-devbox ABS]... [--] <command> [args...]
  -> unseal; merge sealed+adhoc load-devbox Files/Env; materialize under PARENT/<session>/;
     optional home seed+overlay; cwd+SANDBOX_ROOT=session root; HOME=session when home-linked;
     notice: loading devbox per load; exec; cleanup
```

## Preconditions

- Module root is `d.DOCTEST_ROOT/../..` (this tree lives at `tests/sandbox/`).
- `go` is on PATH; `Run(t, d, req)` session-builds `kool` into
  `$TMPDIR/kool-sandbox-doctest-<d.DOCTEST_SESSION_ID>/kool` under a file lock.
  One-process mode: `Run` takes named `d *session.Doctest`; helpers use
  `d.DOCTEST_ROOT` / `d.DOCTEST_SESSION_ID` (no bare free identifiers).
- P1 pack/inspect + baseline sealed run/home-linked are in place. New
  `run/load-devbox/*` leaves are **RED** until runtime `--load-devbox` merge lands.
- Per-leaf isolation: root Setup assigns `WorkingDir = t.TempDir()`.
- Build leaves may take longer (cross-compile); default process timeout is 3m.
- Run leaves use host GOOS/GOARCH and set `KOOL_SANDBOX_ROOT` for materialize.
- Home-linked sealed runs inject fake real home via child `HOME` only
  (`Request.SealedHome` / `SealedEnv`) — never process-global Setenv.
- Load-devbox leaves use `SecondaryPacks` (extra sealed binaries under WorkingDir)
  and `SealedLoadDevbox` / primary `RuntimeLoadDevbox` for load paths.

## Steps

1. Root `Setup` creates an isolated `WorkingDir` and default process timeout.
2. Grouping/leaf `Setup` narrows help/subcommand/flags and writes fixtures under
   `WorkingDir`; run leaves set `AfterBuildRun` + `SealedArgs`; home-linked sets
   `HomeLinked` + `SealedHome` (and optional fake-home files); load-devbox leaves
   set `SecondaryPacks` + `SealedLoadDevbox` and/or `RuntimeLoadDevbox`.
3. `Run` builds/reuses session `kool`, builds `SecondaryPacks` when set, executes
   the primary argv from `Request`, and when `AfterBuildRun` is set, executes the
   sealed binary with `KOOL_SANDBOX_ROOT`, optional `HOME`, and `--load-devbox`
   prefixes from `SealedLoadDevbox`.

## Context

- Shared session cache: `$TMPDIR/kool-sandbox-doctest-<DOCTEST_SESSION_ID>/`
  (`kool` binary + `binaries.ready` + `build.lock`).
- Helpers `writeInputDir`, `writeLocalFile`, `writeFakeHome` prepare config dirs,
  `--file` sources, and home-linked fake real homes under `WorkingDir`.
- No durable product storage; per-leaf temp dirs only.
- Sealed-run capture lives on `Response.Run*` / `Materialize*` / `SealedHome` /
  `SecondaryPaths` fields so P1 leaves keep `ExitCode` as the kool build/help exit.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if err := os.MkdirAll(req.WorkingDir, 0755); err != nil {
		return err
	}
	if req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 3 * time.Minute
	}
	return nil
}

// writeInputDir creates a sandbox input directory layout under WorkingDir.
// name is the directory name (e.g. "in"). files maps sandbox-relative path → content
// (written under name/files/). env maps KEY → value into name/env.yaml. meta is raw
// meta.yaml content (empty skips the file).
func writeInputDir(t *testing.T, workingDir, name string, files map[string]string, env map[string]string, meta string) (string, error) {
	t.Helper()
	root := filepath.Join(workingDir, name)
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", err
	}
	if meta != "" {
		if err := os.WriteFile(filepath.Join(root, "meta.yaml"), []byte(meta), 0644); err != nil {
			return "", err
		}
	}
	if len(files) > 0 {
		for rel, content := range files {
			p := filepath.Join(root, "files", filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
				return "", err
			}
			if err := os.WriteFile(p, []byte(content), 0644); err != nil {
				return "", err
			}
		}
	}
	if len(env) > 0 {
		var b strings.Builder
		// Stable-ish order not required for merge tests; single-key maps in leaves.
		for k, v := range env {
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(v)
			b.WriteString("\n")
		}
		if err := os.WriteFile(filepath.Join(root, "env.yaml"), []byte(b.String()), 0644); err != nil {
			return "", err
		}
	}
	return root, nil
}

// writeLocalFile writes a local file under WorkingDir for --file sources.
func writeLocalFile(t *testing.T, workingDir, rel, content string) (string, error) {
	t.Helper()
	p := filepath.Join(workingDir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		return "", err
	}
	return p, nil
}

// writeFakeHome creates a fake real-home directory under workingDir (name e.g.
// "fake-real-home") and writes files (slash-separated paths relative to that
// home → content). Returns the absolute path of the home directory. Used by
// home-linked leaves so sealed process HOME can point at an isolated tree.
func writeFakeHome(t *testing.T, workingDir, name string, files map[string]string) (string, error) {
	t.Helper()
	root := filepath.Join(workingDir, name)
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", err
	}
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			return "", err
		}
	}
	return root, nil
}
```
