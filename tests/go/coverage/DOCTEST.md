# kool go coverage package-table

`kool go coverage package-table` prints a markdown (or JSON) package-level
coverage table from a classic Go coverprofile. Behavior generalizes
`script/ci/coverage-package-table.py` (doctest reference): parse profile lines,
aggregate by package (directory of file under the module prefix), filter to the
module, apply default skip rules, sort by coverage ascending then package name,
print a table on stdout.

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants

- **User** — invokes `kool go coverage [package-table] [flags] [coverage.out]`.
- **kool go dispatcher** — `tools/go/go_tools.go` routes `coverage` to
  `tools/go/coverage` (implementer wires `case "coverage"`).
- **coverage handler** — root help / unknown subcommand / `package-table`
  flag parse, validation, and table generation.
- **coverprofile parser** — classic `mode:` + `file:line.col,line.col numStmts count`
  lines; pure Go (no xgo).
- **module path resolver** — reads `module` from `go.mod` under `--dir` (default
  `.`) unless `--module` or `--all`.
- **skip policy** — default skip path prefixes `script/`, `cmd/` (relative to
  module) and path substring `/legacy_`; overridable via flags.
- **stdout / stderr writers** — injectable via `HandleWith` for L2 tests.

### Behaviors

- **Coverage root help** — `kool go coverage -h|--help` documents `package-table`
  and principal flags; exit 0; stdout ends with `\n`.
- **package-table help** — `kool go coverage package-table -h|--help` documents
  profile arg and flags (`--module`, `--dir`, `--skip-prefix`, `--skip-contains`,
  `--all`, `--json`); exit 0; trailing `\n`.
- **No subcommand** — bare `kool go coverage` → non-zero; stderr hints
  `package-table` / help.
- **Unknown subcommand** — e.g. `nosuch` → non-zero; stderr indicates unknown /
  unrecognized.
- **Missing profile arg** — `package-table` with no positional → exit **2**;
  usage/error on stderr.
- **Missing profile file** — path not a file → exit **1**; stderr starts with or
  contains `Error:` and mentions the path or "profile".
- **Markdown table (default)** — for matching packages with total statements > 0:

  ```text
  | Coverage | Package |
  |----------|---------|
  | 12.5% | `internal/run` |
  ```

  Coverage = `100 * coveredStmts / totalStmts` formatted `%.1f%%`. Statement
  accounting matches the Python reference: for each profile block, add
  `numStmts` to total; if `count > 0`, also add `numStmts` to covered.
  Package path = directory of the file path relative to the module prefix
  (slash-separated). Sort rows by coverage ascending, then package name
  ascending. Stdout ends with `\n`.
- **Default module filter** — only files under `modulePath + "/"` (from go.mod
  under `--dir`, or `--module`) are counted; other modules omitted.
- **Default skips** — after computing the module-relative path `rel`, skip when
  `rel` has prefix `script/` or `cmd/`, or contains `/legacy_`.
- **`--module PATH`** — use PATH as module prefix (no go.mod read required when
  set).
- **`--all`** — do not filter by module; package path is the directory portion of
  the full cover file path (everything before the last `/`); skip rules still
  apply against that path string where prefixes/contains match.
- **`--skip-prefix LIST`** — comma-separated prefixes relative to module
  (default `script/,cmd/`); replaces the default prefix list when the flag is
  present.
- **`--skip-contains LIST`** — comma-separated substrings (default `/legacy_`);
  replaces the default contains list when the flag is present.
- **Empty after filters** — exit **0**; stdout is header-only table (two lines)
  ending with `\n`; stderr contains `warning:` (case-insensitive) about no
  packages matched.
- **`--json`** — stdout is a JSON array of objects
  `{"coverage": <number>, "package": "<path>"}` sorted the same way; no markdown.
  Numbers are percentage values (e.g. `0`, `50`, `100` or floats). Trailing `\n`.
- **No process cwd mutation** — relative `--dir` / profile paths resolve without
  `Chdir` in product L2 path; tests pass absolute paths.

### Expected implementer API (contract for GREEN)

Package: `github.com/xhd2015/kool/tools/go/coverage`

```go
// Handle is production entry: HandleWith(args, HandleOpts{}).
func Handle(args []string) error

// HandleWith is the injectable entry used by doctests.
// args are the argv after "kool go coverage".
func HandleWith(args []string, opts HandleOpts) error

type HandleOpts struct {
	// Stdout/Stderr nil → os.Stdout / os.Stderr.
	Stdout io.Writer
	Stderr io.Writer
}
```

`tools/go/go_tools.go`: `case "coverage": return coverage.Handle(args)`.

Usage / validation failures that should exit 2 return
`errs.NewSilenceExitCode(2)` after writing usage (or an error that maps to exit 2
via `SilenceExitCode()`). Plain errors map to exit 1 with message on stderr
(prefixed `Error:` for missing profile file).

## Decision Tree

Parameter ranking (most → least significant):

1. **CLI surface** — help / validation / package-table work
2. **Validation kind** — routing vs missing arg vs missing file
3. **Table outcome** — happy aggregate / filter policy / empty / format

```
coverage/
├── help/                                   [usage; exit 0; no profile work]
│   ├── root/                               kool go coverage --help
│   └── package-table/                      package-table --help
├── validation/                             [errors before successful table]
│   ├── no-subcommand/                      bare coverage args
│   ├── unknown-subcommand/                 nosuch
│   ├── missing-arg/                        package-table without profile → exit 2
│   └── missing-file/                       profile path not a file → exit 1 Error:
└── package-table/                          [parse + filter + render]
    ├── basic-sorted/                       two packages → markdown ascending
    ├── default-skips/                      script/, cmd/, /legacy_ omitted
    ├── module-from-dir/                    --dir go.mod; outside module omitted
    ├── module-override/                    --module explicit prefix
    ├── all-modules/                        --all keeps foreign module packages
    ├── empty-match/                        warning + header-only; exit 0
    ├── custom-skips/                       --skip-prefix / --skip-contains replace defaults
    └── json/                               --json stable row shape
```

## Test Index

| Leaf | Description | Expect until implement |
|------|-------------|------------------------|
| `help/root/` | Coverage `--help` exit 0; mentions package-table; trailing `\n` | RED |
| `help/package-table/` | package-table `--help` documents flags + profile; trailing `\n` | RED |
| `validation/no-subcommand/` | Bare coverage → non-zero; stderr hints package-table/help | RED |
| `validation/unknown-subcommand/` | Unknown subcommand → non-zero | RED |
| `validation/missing-arg/` | No profile positional → exit 2 | RED |
| `validation/missing-file/` | Missing file → exit 1; `Error:` on stderr | RED |
| `package-table/basic-sorted/` | Two packages 0% and 100% → sorted markdown | RED |
| `package-table/default-skips/` | script/, cmd/, legacy_ rows omitted | RED |
| `package-table/module-from-dir/` | go.mod via `--dir`; foreign module omitted | RED |
| `package-table/module-override/` | `--module` filters without relying on other module path | RED |
| `package-table/all-modules/` | `--all` includes foreign module package | RED |
| `package-table/empty-match/` | All filtered out → warning + headers; exit 0 | RED |
| `package-table/custom-skips/` | Custom skip lists replace defaults | RED |
| `package-table/json/` | `--json` array of `{coverage,package}` sorted | RED |

## How to Run

```sh
doctest vet ./tests/go/coverage
doctest test ./tests/go/coverage
```

Classic TDD: tree is written first. Until `tools/go/coverage` exists (and
`go_tools.go` wires `coverage`), `doctest test` is **RED** (compile and/or
assertion failure). Prefer L2 in-process `HandleWith` — no kool binary e2e.

```go
import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/kool/pkgs/errs"
	gocoverage "github.com/xhd2015/kool/tools/go/coverage"
)

// Request drives one in-process tools/go/coverage.HandleWith invocation.
// Args are relative to `kool go coverage` (subcommand + flags + profile).
type Request struct {
	// Args is the full argv after "coverage". When non-nil (including empty
	// slice), Run uses Args and ignores the structured fields below except
	// WorkingDir.
	Args []string

	// Structured builders (used when Args == nil).
	// HelpAtRoot: coverage --help
	HelpAtRoot bool
	// HelpPackageTable: package-table --help
	HelpPackageTable bool
	// Subcommand is the first positional (e.g. "package-table", "nosuch").
	// Empty with HelpAtRoot false = bare coverage (no subcommand).
	Subcommand string

	// ProfilePath is the coverage.out positional (absolute path in tests).
	ProfilePath string
	// ProfileSet forces passing ProfilePath even when empty.
	ProfileSet bool

	// Flags for package-table.
	Module       string
	ModuleSet    bool
	Dir          string
	DirSet       bool
	SkipPrefix   string
	SkipPrefixSet bool
	SkipContains string
	SkipContainsSet bool
	All          bool
	JSON         bool

	// WorkingDir is isolation root for fixtures (absolute). Set by root Setup.
	// Not process cwd — leaves write files here and pass absolute --dir/profile.
	WorkingDir string
}

// Response is CLI capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func buildArgs(req *Request) []string {
	if req.Args != nil {
		return append([]string(nil), req.Args...)
	}
	if req.HelpAtRoot {
		return []string{"--help"}
	}
	if req.HelpPackageTable {
		return []string{"package-table", "--help"}
	}
	var args []string
	if req.Subcommand != "" {
		args = append(args, req.Subcommand)
	}
	if req.Subcommand == "package-table" {
		if req.ModuleSet {
			args = append(args, "--module", req.Module)
		}
		if req.DirSet {
			args = append(args, "--dir", req.Dir)
		}
		if req.SkipPrefixSet {
			args = append(args, "--skip-prefix", req.SkipPrefix)
		}
		if req.SkipContainsSet {
			args = append(args, "--skip-contains", req.SkipContains)
		}
		if req.All {
			args = append(args, "--all")
		}
		if req.JSON {
			args = append(args, "--json")
		}
		if req.ProfileSet || req.ProfilePath != "" {
			args = append(args, req.ProfilePath)
		}
	}
	return args
}

func mapExit(err error) (int, string) {
	if err == nil {
		return 0, ""
	}
	if se, ok := errs.IsSilenceExitCode(err); ok {
		return se.SilenceExitCode(), ""
	}
	var exitAware interface{ SilenceExitCode() int }
	if errors.As(err, &exitAware) {
		return exitAware.SilenceExitCode(), ""
	}
	// Match main.go: print error to stderr and exit 1.
	return 1, err.Error()
}

// Run invokes tools/go/coverage.HandleWith with captured writers.
// L2 in-process only — never shells the kool binary.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	var stdout, stderr bytes.Buffer
	resp := &Response{}

	opts := gocoverage.HandleOpts{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	runErr := gocoverage.HandleWith(buildArgs(req), opts)
	code, errMsg := mapExit(runErr)
	if errMsg != "" {
		// Mirror main.go when handler returns a plain error without writing stderr.
		if stderr.Len() == 0 || !strings.Contains(stderr.String(), errMsg) {
			// Prefer Error: prefix if the message does not already have it.
			if strings.HasPrefix(errMsg, "Error:") {
				fmt.Fprintln(&stderr, errMsg)
			} else {
				fmt.Fprintln(&stderr, "Error:", errMsg)
			}
		}
	}
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.ExitCode = code
	return resp, nil
}

// writeFile creates parent dirs and writes content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// writeGoMod writes a minimal go.mod at dir for modulePath.
func writeGoMod(t *testing.T, dir, modulePath string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
}

// writeCoverProfile writes a classic coverprofile at path.
func writeCoverProfile(t *testing.T, path, body string) {
	t.Helper()
	if !strings.HasPrefix(body, "mode:") {
		body = "mode: set\n" + body
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	writeFile(t, path, body)
}

// ensure writers compile against io.Writer.
var _ io.Writer = (*bytes.Buffer)(nil)
```
