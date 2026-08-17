# kool with-go — injectable resolve / list / exec wrap

L2 in-process doctests for `github.com/xhd2015/kool/tools/go/with_go`.
`Run` calls `HandleWith`, `ResolveGorootWith`, `ListWith`, and `ExecGoroot`.
No `kool` binary, no network, no `go run download-go`.

**Classic TDD:** the injectable wrap (`HandleWith`, `ResolveGorootWith`,
`ListWith`) does not exist yet. The suite is **compile-RED** until the
implementer lands those funcs. Existing `Handle` / `ResolveGoroot` /
`InstallGo` (`go run download-go`) are the old surface — leaves must not
call only those old signatures.

# DSN (Domain Specific Notion)

kool `with-go` pins a go version, resolves `$InstallDir/<pin>`, lists
downloadable versions, and execs a child under that GOROOT. Callers inject
install dir, install hook, writers, and `FetchHTML`.

**Participants**

- **`HandleWith`** — CLI dispatch. Empty args error (`kool with-go`).
  `list` → `ListWith`. `GOROOT=<path>` skips resolve. `go1.19` / `1.19`
  → `ResolveGorootWith` then `ExecGoroot`.
- **`ResolveGorootWith`** — pin via `withgo.PinPatch`, dest
  `$InstallDir/<pin>`. Existing dest directory is returned; Install and
  Prompt unused. Missing + `Download: false` is an error. Missing +
  `Download: true` writes Prompt to Stderr then calls Install (or
  `downloadgo.Download`).
- **`ListWith`** — `downloadgo.List` with injected `FetchHTML`; writes
  `go%s\n` per naked version to `w`.
- **`ExecGoroot`** — kool-owned child: `GOROOT=$abs`,
  `PATH=$abs/bin:$PATH`. Bare `go` becomes `$GOROOT/bin/go` when present.
  Empty args run `env`.
- **Install hook / writers / InstallDir / FetchHTML** — test seams.
  Leaves pass `t.TempDir()` and recording hooks; never real network.

**Behaviors**

- Pin: `go1.19` / `1.19` → dest suffix `go1.19.13`.
- Dest exists as a directory → return it; hook unused; no Prompt.
- Missing + `Download: false` → error; hook unused.
- Missing + `Download: true` → Prompt on Stderr if set, then Install(pin, dir).
- `GOROOT=` skips resolve; Install unused.
- List stdout is `go` + naked version, one per line.
- Exec child sees absolute GOROOT and PATH0 `$GOROOT/bin`.
- No process-global `Setenv` / `Chdir` / stdio rewrite.

## Version

0.0.2

## Decision Tree

Root splits on **operation** (`req.Op`). Under resolve, dest existence;
under dest-missing, `Download`; under handle, arg class.

```
tests/go/with-go/                       [op]
├── resolve/                            [dest existence]
│   ├── dest-present/                   $InstallDir/go1.19.13 is a directory
│   │   ├── return-path/                go1.19 → that path; hook unused; no Prompt
│   │   └── pin-suffix/                 1.19 → path ends with go1.19.13
│   └── dest-missing/                   dest absent
│       ├── download-false/             error; hook unused
│       └── download-true/              hook(pin=go1.19.13); Prompt on Stderr
├── handle/                             [arg class]
│   ├── empty-args/                     error containing kool with-go
│   ├── goroot-skip-resolve/            GOROOT=/abs/fake + /usr/bin/true; hook unused
│   └── version-then-exec/              go1.19 dest-exists + fake go → ExecGoroot
├── list/                               [injected HTML]
│   └── injected-html/                  go… lines; FetchHTML used; no network
└── exec/                               [fake $GOROOT/bin/go]
    └── bare-go/                        child GOROOT abs; PATH0=$goroot/bin
```

### Parameter significance (high → low)

1. **Operation** — resolve / handle / list / exec.
2. **Dest existence** (resolve) or **arg class** (handle).
3. **Download / Prompt / version spelling** — only when the parent uses them.

## Test Index

| # | Leaf | Description | Expected |
|---|------|-------------|----------|
| 1 | `resolve/dest-present/return-path` | dest dir already at `$InstallDir/go1.19.13` | that path; hook unused; no Prompt |
| 2 | `resolve/dest-present/pin-suffix` | naked `1.19` dest-exists fixture | path ends with `go1.19.13` |
| 3 | `resolve/dest-missing/download-false` | missing + `Download: false` | error; hook unused |
| 4 | `resolve/dest-missing/download-true` | missing + hook + Prompt | hook(`go1.19.13`, InstallDir); Prompt on Stderr |
| 5 | `handle/empty-args` | no args | error containing `kool with-go` |
| 6 | `handle/goroot-skip-resolve` | `GOROOT=/abs/fake` + `/usr/bin/true` | hook count 0 |
| 7 | `handle/version-then-exec` | `go1.19` + dest-exists + fake `bin/go` | resolve then exec; hook unused |
| 8 | `list/injected-html` | injected FetchHTML | `go…` lines; fetch used |
| 9 | `exec/bare-go` | fake `$GOROOT/bin/go` | child GOROOT abs; PATH0=`$goroot/bin` |

## How to Run

From the kool module root:

```sh
doctest vet ./tests/go/with-go
doctest test ./tests/go/with-go
```

Classic TDD: `doctest vet` must pass; `doctest test` is **compile-RED**
until `HandleWith` / `ResolveGorootWith` / `ListWith` exist. Do not add
a `go.mod` replace here — the implementer does that.

```go
import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/kool/tools/go/with_go"
	"github.com/xhd2015/xgo/support/downloadgo"
)

type Request struct {
	Op string // resolve | handle | list | exec

	GoVersion string

	Args []string
	Envs []string

	InstallDir    string
	Download      bool
	Prompt        string
	RecordInstall bool
	HookGoroot    string

	Goroot string // exec: fake GOROOT

	FetchHTML   func(ctx context.Context) (string, error)
	RecordFetch bool
}

type Response struct {
	Goroot      string
	Stdout      string
	Stderr      string
	HookCalled  bool
	HookVersion string
	HookDir     string
	HookCount   int
	FetchCount  int
	Err         error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	resp := &Response{}
	var stdout, stderr bytes.Buffer

	switch req.Op {
	case "resolve":
		goroot, err := with_go.ResolveGorootWith(req.GoVersion, resolveOpts(req, resp, &stdout, &stderr))
		resp.Goroot = goroot
		resp.Err = err
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()

	case "handle":
		resp.Err = with_go.HandleWith(req.Args, req.Envs, resolveOpts(req, resp, &stdout, &stderr))
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if out := readChildOut(req); out != "" {
			resp.Stdout = out
		}

	case "list":
		listOpts := downloadgo.ListOptions{}
		if req.RecordFetch || req.FetchHTML != nil {
			fetch := req.FetchHTML
			listOpts.FetchHTML = func(ctx context.Context) (string, error) {
				resp.FetchCount++
				if fetch == nil {
					return "", fmt.Errorf("FetchHTML not injected")
				}
				return fetch(ctx)
			}
		}
		resp.Err = with_go.ListWith(context.Background(), listOpts, &stdout)
		resp.Stdout = stdout.String()

	case "exec":
		resp.Err = with_go.ExecGoroot(req.Goroot, req.Args, req.Envs)
		resp.Stdout = readChildOutPath(req.Goroot)

	default:
		return nil, fmt.Errorf("unknown op: %s", req.Op)
	}

	return resp, nil
}

func resolveOpts(req *Request, resp *Response, stdout, stderr *bytes.Buffer) withgo.ResolveOptions {
	opts := withgo.ResolveOptions{
		InstallDir: req.InstallDir,
		Download:   req.Download,
		Prompt:     req.Prompt,
		Stdout:     stdout,
		Stderr:     stderr,
	}
	if req.RecordInstall {
		opts.Install = func(ctx context.Context, version, installDir string) (string, error) {
			resp.HookCalled = true
			resp.HookCount++
			resp.HookVersion = version
			resp.HookDir = installDir
			if req.HookGoroot != "" {
				return req.HookGoroot, nil
			}
			return filepath.Join(installDir, version), nil
		}
	}
	return opts
}

func readChildOutPath(goroot string) string {
	if goroot == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(goroot, "child.out"))
	if err != nil {
		return ""
	}
	return string(b)
}

func readChildOut(req *Request) string {
	var dirs []string
	if req.Goroot != "" {
		dirs = append(dirs, req.Goroot)
	}
	if req.InstallDir != "" {
		dirs = append(dirs, destPin(req.InstallDir))
	}
	if req.HookGoroot != "" {
		dirs = append(dirs, req.HookGoroot)
	}
	for _, dir := range dirs {
		if out := readChildOutPath(dir); out != "" {
			return out
		}
	}
	return ""
}
```
