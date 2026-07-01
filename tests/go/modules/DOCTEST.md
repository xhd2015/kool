# kool go modules --list

`kool go modules --list [--dir <root>]` streams `<dir> <module-path>` lines to stdout, one
per Go module found, in walk order (unsorted). `dir` is `.` for the root module and the
plain slash-relative path (e.g. `sub-dir`, `nested/service`) for sub-directories — no `./`
prefix. The walk delegates to `github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan` and
applies the same skip rules (`.git`/`vendor`/`testdata`, gitignored, nested separate repos).

The `--list` flag is mutually exclusive with `ls-files` / `update-local-deps` subcommands
and with `--no-tags` (tags are not part of list output). Default behavior (no `--list`) is
the existing tree render, unchanged.

# DSN (Domain Specific Notion)

The **user** runs `kool go modules --list --dir <root>` from anywhere. The **handler** sets
up the scan package's `Options`, calls `scan.ScanStream(root, opts, fn)`, and for each
`Module` emitted writes `<dir> <path>\n` to stdout, flushing per line. The `<dir>` field is
the module's `Dir` (`.`, or slash-relative, no `./`); `<path>` is the go.mod module path.
Output is walk-order (the stream emits as the walker discovers) — sorting would require
buffering the whole walk, so `--list` is intentionally unsorted.

Skip rules are the scan package's: name skips (`.git`/`vendor`/`testdata`), and when the
root is a git repo, gitignored dirs and nested separate repos (own `.git`, not a submodule).
A nested separate repo's own `go.mod` is never emitted.

## Version

0.0.2

## Decision Tree

The single factor is the **workspace layout** relative to the skip rules — that is what
changes which lines appear in `--list` output. `list/` is the only operation mode exercised
here (the `--list` flag is on); siblings are MECE over the layout: a clean multi-module
workspace (basic), a workspace with a nested separate repo (must be absent), and a workspace
with `testdata` (must be absent).

```
modules tests
└── list/
    ├── basic/                # root + sub-dir -> both emitted, space-separated, walk order
    ├── skips-nested-repo/    # ext/ own .git, not submodule -> ext line absent
    └── skips-testdata/       # testdata/ pruned -> testdata line absent
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `list/basic` | root + `sub-dir` -> stdout has `. some.com/root` and `sub-dir some.com/root/sub` |
| 2 | `list/skips-nested-repo` | `ext/` nested separate repo -> no `ext` line in stdout |
| 3 | `list/skips-testdata` | `testdata/` pruned -> no `testdata` line in stdout |

## How to Run

```sh
# requires `kool` built with the local replace directive (see SETUP.md preconditions)
doctest vet ./tests/go/modules
doctest test ./tests/go/modules
```

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

type Request struct {
	RootDir string // populated by leaf Setup: workspace root passed via --dir
}

type Response struct {
	Err      error
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run execs `kool go modules --list --dir <req.RootDir>` and captures stdout/stderr/exit.
// Requires `kool` on PATH, rebuilt with the local replace directive pointing at the new
// (unpublished) scan package. Leaves assert on resp.Stdout line set.
func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}

	if _, err := exec.LookPath("kool"); err != nil {
		return nil, fmt.Errorf("kool not found in PATH, build it first with the local replace: %w", err)
	}

	cmd := exec.Command("kool", "go", "modules", "--list", "--dir", req.RootDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
			resp.Err = runErr
		} else {
			return nil, fmt.Errorf("failed to run kool: %w", runErr)
		}
	}

	return resp, nil
}

var _ = os.Getwd
```
