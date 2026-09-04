# kool go modcache

`kool go modcache inspect` reports versions in `$GOMODCACHE`. Non-newest versions
of each module path are **legacy**. `kool go modcache prune` deletes those
legacy copies (extracted dir + matching download files). `--dry-run` prints the
plan and does not delete. `golang.org/toolchain` is excluded unless
`--include-toolchain`. Optional `--root` uses git-repo discovery (same as
`kool git scan-repos`) to hint local go.mod/go.sum still on legacy versions.

## Version

0.0.3

## DSN (Domain Specific Notion)

### Participants

- **User** — invokes `kool go modcache <inspect|prune> [flags]`.
- **kool go dispatcher** — `tools/go/go_tools.go` routes `modcache`.
- **modcache handler** — `tools/go/modcache.HandleWith` (injectable stdout/stderr).
- **inventory** — walks extracted `<path>@<version>` dirs and `cache/download/.../@v/` files.
- **live set** — optional `--root` → `scan_repo.Scan` + `gotool/mod/scan` + go.sum.

### Behaviors

- **Root help** — `kool go modcache -h|--help|help` documents inspect and prune; exit 0; trailing `\n`.
- **Subcommand help** — `inspect --help` / `prune --help`; exit 0; trailing `\n`.
- **No subcommand** — non-zero; stderr hints inspect/prune/help.
- **Unknown subcommand** — non-zero; stderr `Error:` and unknown/unrecognized.
- **Missing modcache** — `--modcache` path not a directory → exit 1; `Error:` on stderr.
- **inspect** — summary of totals, module counts, legacy; `SAVE:` estimated bytes if prune keeps newest of each module (`N% of total` when save > 0); TOP table of legacy paths; never deletes.
- **Progress** — while sizing, stderr gets flush-left `[n/total] kind msg` stages (`extracted`, `download`, `vcs`, and `live` when `--root`). Kind-aligned `i/N path@version` heartbeats every 25 extracted versions (plus first and last). Report stays on stdout.
- **inspect --json** — `Report` JSON (modules, `legacyBytes`, `saveBytes` equal to `legacyBytes`, suggestions); no ANSI; stage markers still on stderr.
- **inspect --root** — upgrade suggestions when a local require is older than cache-newest.
- **prune --dry-run** — `would remove N versions, SIZE` plus `rm` lines; files remain.
- **prune** — deletes legacy; keeps newest; keeps toolchain without `--include-toolchain`.
- **No process cwd/env mutation** — tests pass absolute `--modcache` / `--root`.

### Expected implementer API

Package: `github.com/xhd2015/kool/tools/go/modcache`

```go
func Handle(args []string) error
func HandleWith(args []string, opts HandleOpts) error
type HandleOpts struct {
	Stdout io.Writer
	Stderr io.Writer
}
```

`tools/go/go_tools.go`: `case "modcache": return modcache.Handle(args)`.

## Decision Tree

1. **CLI surface** — help / validation / inspect / prune
2. **inspect fixture** — empty / one version / two versions / encoding / toolchain / live-set
3. **prune effect** — dry-run vs delete (toolchain preserved)

```
modcache/
├── help/
│   ├── root/
│   ├── inspect/
│   └── prune/
├── validation/
│   ├── no-subcommand/
│   ├── unknown-subcommand/
│   └── missing-modcache/
├── inspect/
│   ├── empty/
│   ├── single-version/
│   ├── multi-version/
│   ├── escape-path/
│   ├── toolchain-separate/
│   ├── json/
│   ├── progress/
│   └── with-root/
└── prune/
    ├── dry-run/
    └── apply/
```

Intentional exclusions: `--include-toolchain` prune (flag exists; default skip is the user-visible contract); `cache/vcs` deletion; network “true latest”.

## Test Index

| Leaf | Description |
|------|-------------|
| `help/root/` | Root `--help` mentions inspect and prune; trailing `\n` |
| `help/inspect/` | `inspect --help` documents `--modcache` |
| `help/prune/` | `prune --help` documents `--dry-run` |
| `validation/no-subcommand/` | Bare modcache → non-zero; stderr hints inspect/prune |
| `validation/unknown-subcommand/` | Unknown subcommand → `Error:` |
| `validation/missing-modcache/` | Missing `--modcache` dir → exit 1 `Error:` |
| `inspect/empty/` | Empty cache → 0 paths, 0 legacy, SAVE 0B |
| `inspect/single-version/` | One version → not listed as legacy; SAVE 0B |
| `inspect/multi-version/` | Older version is legacy; KEEP is newest; SAVE > 0 with % of total |
| `inspect/escape-path/` | `!azure` reports as `github.com/Azure/...` |
| `inspect/toolchain-separate/` | toolchain not in default legacy |
| `inspect/json/` | `--json` Report shape with newest + legacy; `saveBytes` == `legacyBytes`; markers on stderr only |
| `inspect/progress/` | stderr `[1/3] extracted` / download / vcs; stdout is the report |
| `inspect/with-root/` | local require older than cache-newest → suggestion; `[4/4] live` |
| `prune/dry-run/` | would remove; files remain |
| `prune/apply/` | older gone, newest kept, toolchain kept |

## How to Run

```sh
doctest vet ./tests/go/modcache
doctest test ./tests/go/modcache
```

L2 in-process `HandleWith` — no kool binary e2e.

```go
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/kool/pkgs/errs"
	"github.com/xhd2015/kool/tools/go/modcache"
	"golang.org/x/mod/module"
)

type Request struct {
	Args []string

	HelpAtRoot    bool
	HelpInspect   bool
	HelpPrune     bool
	Subcommand    string

	ModCache    string
	ModCacheSet bool
	Roots       []string
	JSON        bool
	DryRun      bool
	IncludeToolchain bool
	NoCache     bool
	CacheDir    string

	WorkingDir string
}

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
	if req.HelpInspect {
		return []string{"inspect", "--help"}
	}
	if req.HelpPrune {
		return []string{"prune", "--help"}
	}
	var args []string
	if req.Subcommand != "" {
		args = append(args, req.Subcommand)
	}
	if req.ModCacheSet || req.ModCache != "" {
		args = append(args, "--modcache", req.ModCache)
	}
	for _, r := range req.Roots {
		args = append(args, "--root", r)
	}
	if req.JSON {
		args = append(args, "--json")
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	if req.IncludeToolchain {
		args = append(args, "--include-toolchain")
	}
	if req.NoCache {
		args = append(args, "--no-cache")
	}
	if req.CacheDir != "" {
		args = append(args, "--cache-dir", req.CacheDir)
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
	return 1, err.Error()
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	var stdout, stderr bytes.Buffer
	resp := &Response{}
	runErr := modcache.HandleWith(buildArgs(req), modcache.HandleOpts{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	code, errMsg := mapExit(runErr)
	if errMsg != "" {
		if stderr.Len() == 0 || !strings.Contains(stderr.String(), errMsg) {
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeExtracted(t *testing.T, modcacheDir, path, version, content string) string {
	t.Helper()
	escaped, err := module.EscapePath(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(modcacheDir, filepath.FromSlash(escaped)+"@"+version)
	writeFile(t, filepath.Join(dir, "x.txt"), content)
	return dir
}

func writeZip(t *testing.T, modcacheDir, path, version, content string) string {
	t.Helper()
	escaped, err := module.EscapePath(path)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(modcacheDir, "cache", "download", filepath.FromSlash(escaped), "@v", version+".zip")
	writeFile(t, p, content)
	return p
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test User")
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "init")
}

var _ = json.Marshal
```
