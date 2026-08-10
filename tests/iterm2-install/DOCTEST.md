# kool iterm2 install — `--via-open` (user-driven Gatekeeper path)

`kool iterm2 install --via-open` maps the CLI flag onto
`iterm2install.InstallOpts.InstallViaUserOpen = true` so the official zip install
finishes via clear-quarantine + user `open` instead of the fully automated path.

**Default layer: L2** in-process `tools/iterm2.RunForTest` with injectable fake
HTTP for resolve/dry-run. No real network, no real `open` / `xattr`.

**Classic TDD:** `--via-open` is not implemented yet. Expect **RED** (flag
rejected, help missing, or mode line absent) until the implementer lands the
flag + dry-run wording + conflict check.

**Out of scope (P1):** codex command, changing dot-pkgs library, real download
e2e, publishing.

## Version

0.0.1

## DSN (Domain Specific Notion)

### Participants

- **User** — invokes `kool iterm2 install [flags]`.
- **kool CLI router** — `tools/iterm2` dispatches reserved first arg `install` to
  `runInstall`.
- **`runInstall` / `RunForTest`** — parses flags (`--dry-run`, `--download-only`,
  `--download-dir`, **`--via-open`**), validates conflicts, resolves latest URL,
  prints dry-run plan or performs download/install.
- **Fake HTTP latest endpoint** — L2 inject of `installLatestURL` +
  `installHTTPClient` (package vars; exposed for doctests via
  `SetInstallHTTPForTest`) so resolve never hits iterm2.com.
- **Library `shell/iterm2/install`** — `InstallOpts.InstallViaUserOpen` (bool,
  default false); open/clear hooks live in the library (out of this suite’s
  direct assert surface).

### Behaviors

- **Help** — `kool iterm2 install --help` documents `--via-open` and mentions
  open / Gatekeeper / user-driven install; exit 0; stdout ends with `\n`.
- **`--via-open` default off** — dry-run without the flag does **not** claim
  via-open / user-open mode.
- **`--dry-run --via-open`** — resolve via fake HTTP; stdout dry-run banner +
  version; plan/steps mention open (or clear-quarantine / via-open mode);
  exit 0; **no zip written**.
- **`--via-open` + `--download-only`** — hard error: non-zero exit, stderr
  starts with or contains `Error:`, no download side effects.
- **Open/install failure** — strict exit 1 (library path; out of suite for P1
  dry-run/validation leaves).
- **SkipScriptable** remains false (kool default) on full install; dry-run does
  not change that contract.
- **Flag name** exactly `--via-open`.

### Expected implementer surface (for GREEN)

Package: `github.com/xhd2015/kool/tools/iterm2`

- Parse `--via-open` in `runInstall`; set
  `InstallOpts.InstallViaUserOpen = true` when flag set (and not download-only).
- Reject `--via-open` together with `--download-only` before download.
- Dry-run plan reflects via-open mode (mode line and/or steps).
- Export L2 inject helper used by this tree (wrapper over existing package vars):

```go
// SetInstallHTTPForTest injects latest URL + HTTP client for install resolve/download.
// Returns restore; Run always restores (also safe under a process mutex in this suite).
func SetInstallHTTPForTest(latestURL string, client *http.Client) (restore func())
```

`RunForTest(args, stdout, stderr, workingDir)` already exists; pass
`workingDir=""` (no chdir).

## Decision Tree

```
iterm2-install/
├── DOCTEST.md
├── SETUP.md
├── help/                                   [usage; exit 0]
│   ├── SETUP.md
│   └── show-usage/                         --help lists --via-open + open/Gatekeeper
├── dry-run/                                [fake HTTP resolve; no zip]
│   ├── SETUP.md
│   ├── via-open/                           --dry-run --via-open → mode/steps mention open
│   └── default-no-via-open/                --dry-run without flag → no via-open claim
└── validation/                             [hard errors before download]
    ├── SETUP.md
    └── via-open-with-download-only/        both flags → Error:; exit 1; no zip
```

### Parameter significance (high → low)

1. **Outcome class** — help | dry-run plan | validation error
2. **`--via-open` on/off** — maps InstallViaUserOpen; conflicts with download-only
3. **HTTP inject** — required for resolve paths; help/validation-early may skip

## Test Index

| Leaf | Description | Classic |
|------|-------------|---------|
| `help/show-usage/` | `--help` exit 0; includes `--via-open` + open/Gatekeeper/user-driven; trailing `\n` | RED |
| `dry-run/via-open/` | `--dry-run --via-open` + fake HTTP → banner, version, open/via-open steps; exit 0; no zip | RED |
| `dry-run/default-no-via-open/` | `--dry-run` only → success; stdout must **not** claim via-open mode | RED* |
| `validation/via-open-with-download-only/` | both flags → `Error:` on stderr; non-zero; no zip | RED |

\*default-no-via-open may already pass dry-run success against current product;
via-open absence assert stays valid; still seal with suite for regression.

## How to Run

```sh
# from kool module root (external/kool-master-2026-08-10-1)
doctest vet ./tests/iterm2-install
doctest test ./tests/iterm2-install
doctest test -v ./tests/iterm2-install/help/show-usage
doctest test -v ./tests/iterm2-install/dry-run/via-open
doctest test -v ./tests/iterm2-install/validation/via-open-with-download-only
```

Classic TDD: expect **RED** until `--via-open` lands. Package HTTP inject vars
are process-global — this harness serializes inject + `RunForTest` with a mutex
(do not rely on leaf `t.Parallel()` isolation for those vars).

```go
import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

// installInjectMu serializes package-level installHTTPClient / installLatestURL
// mutation (sequential-sensitive inject surface).
var installInjectMu sync.Mutex

// Request drives one in-process kool iterm2 install invocation via RunForTest.
type Request struct {
	// Help: install --help (ignores other flags except WorkingDir).
	Help bool

	// Flags (install subcommand).
	DryRun       bool
	ViaOpen      bool
	DownloadOnly bool
	// DownloadDir is --download-dir when non-empty (leaf TempDir for isolation).
	DownloadDir string

	// UseFakeHTTP starts a per-leaf httptest latest→zip fixture and injects
	// SetInstallHTTPForTest. Required for dry-run leaves that resolve.
	UseFakeHTTP bool
	// FinalZipName is the redirect target basename (default iTerm2-3_6_11.zip).
	FinalZipName string

	// WorkingDir is unused by RunForTest for install (pass ""); reserved.
	WorkingDir string
}

// Response is CLI capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	// ZipPath is the planned zip under DownloadDir (when set); Assert may Stat it.
	ZipPath string
}

func buildInstallArgs(req *Request) []string {
	args := []string{"install"}
	if req.Help {
		args = append(args, "--help")
		return args
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	if req.ViaOpen {
		args = append(args, "--via-open")
	}
	if req.DownloadOnly {
		args = append(args, "--download-only")
	}
	if req.DownloadDir != "" {
		args = append(args, "--download-dir", req.DownloadDir)
	}
	return args
}

func startFakeLatestServer(t *testing.T, zipName string) *httptest.Server {
	t.Helper()
	if zipName == "" {
		zipName = "iTerm2-3_6_11.zip"
	}
	// Minimal non-empty body (dry-run must not download; validation must not either).
	body := []byte("PK\x03\x04fake-iterm2-zip")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/latest"):
			http.Redirect(w, r, "http://"+r.Host+"/"+zipName, http.StatusFound)
		case strings.HasSuffix(r.URL.Path, "/"+zipName) || strings.HasSuffix(r.URL.Path, zipName):
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// Run invokes tools/iterm2.RunForTest for install. When UseFakeHTTP, injects
// fake latest endpoint under installInjectMu + SetInstallHTTPForTest.
// workingDir is always "" (no t.Chdir / product chdir for install leaves).
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	zipName := req.FinalZipName
	if zipName == "" {
		zipName = "iTerm2-3_6_11.zip"
	}
	if req.DownloadDir != "" {
		if err := os.MkdirAll(req.DownloadDir, 0o755); err != nil {
			return nil, err
		}
	}

	resp := &Response{}
	if req.DownloadDir != "" {
		resp.ZipPath = filepath.Join(req.DownloadDir, zipName)
	}

	args := buildInstallArgs(req)
	var stdout, stderr bytes.Buffer

	runOnce := func() {
		code := iterm2.RunForTest(args, &stdout, &stderr, "")
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = code
	}

	if req.UseFakeHTTP {
		srv := startFakeLatestServer(t, zipName)
		installInjectMu.Lock()
		defer installInjectMu.Unlock()
		restore := iterm2.SetInstallHTTPForTest(srv.URL+"/latest", srv.Client())
		defer restore()
		runOnce()
		return resp, nil
	}

	// Help / early validation: no HTTP inject. Still serialize in case other
	// parallel leaves hold inject (cheap).
	installInjectMu.Lock()
	defer installInjectMu.Unlock()
	runOnce()
	return resp, nil
}

// zipExists reports whether path exists as a regular file.
func zipExists(path string) bool {
	if path == "" {
		return false
	}
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// assertNoZip fails if DownloadDir zip was written.
func assertNoZip(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ZipPath == "" {
		return
	}
	if zipExists(resp.ZipPath) {
		t.Fatalf("expected no zip written at %s", resp.ZipPath)
	}
}

// silence unused in some leaves
var _ = fmt.Sprintf
```
