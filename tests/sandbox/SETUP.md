# Scenario

**Feature**: kool sandbox build + sealed-binary run (P1 package, P2 unpack/exec)

```
# help
user -> kool sandbox [--help | build --help]
  -> usage on stdout, exit 0

# validation
user -> kool sandbox build [missing -o | empty pack | bad --env | missing --file source]
  -> stderr error, non-zero; no sealed artifact required

# build happy
user -> kool sandbox build -o OUT [-i DIR] [--file L=R]... [--env K=V]... [--goos] [--goarch]
  -> one-time RSA + AES-GCM sealed PackBlob embedded in OUT; summary on stdout

# inspect (optional post-build)
user -> kool sandbox inspect OUT
  -> paths, content hashes, env keys (never secret values)

# sealed run (P2)
user -> KOOL_SANDBOX_ROOT=PARENT ./OUT [--] <command> [args...]
  -> unseal; materialize under PARENT/<session>/; cwd+SANDBOX_ROOT=session root;
     exec command; exit=child; remove session dir on exit
```

## Preconditions

- Module root is `d.DOCTEST_ROOT/../..` (this tree lives at `tests/sandbox/`).
- `go` is on PATH; `Run(t, d, req)` session-builds `kool` into
  `$TMPDIR/kool-sandbox-doctest-<d.DOCTEST_SESSION_ID>/kool` under a file lock.
  One-process mode: `Run` takes named `d *session.Doctest`; helpers use
  `d.DOCTEST_ROOT` / `d.DOCTEST_SESSION_ID` (no bare free identifiers).
- P1 (`build/*`, `help/*`) is implemented; P2 sealed runner may still stub with
  `Error: run not implemented` — `run/*` leaves expected **RED** until implementer.
- Per-leaf isolation: root Setup assigns `WorkingDir = t.TempDir()`.
- Build leaves may take longer (cross-compile); default process timeout is 3m.
- Run leaves use host GOOS/GOARCH and set `KOOL_SANDBOX_ROOT` for materialize.

## Steps

1. Root `Setup` creates an isolated `WorkingDir` and default process timeout.
2. Grouping/leaf `Setup` narrows help/subcommand/flags and writes fixtures under
   `WorkingDir`; run leaves set `AfterBuildRun` + `SealedArgs`.
3. `Run` builds/reuses session `kool`, executes the argv from `Request`, and when
   `AfterBuildRun` is set, executes the sealed binary with `KOOL_SANDBOX_ROOT`.

## Context

- Shared session cache: `$TMPDIR/kool-sandbox-doctest-<DOCTEST_SESSION_ID>/`
  (`kool` binary + `binaries.ready` + `build.lock`).
- Helpers `writeInputDir`, `writeLocalFile` prepare config dirs and `--file`
  sources under `WorkingDir`.
- No durable product storage; per-leaf temp dirs only.
- Sealed-run capture lives on `Response.Run*` / `Materialize*` fields so P1
  leaves keep `ExitCode` as the kool build/help exit.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
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
```
