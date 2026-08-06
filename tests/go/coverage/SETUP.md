# Scenario

**Feature**: kool go coverage package-table — coverprofile → package coverage table

```
# help
user -> kool go coverage [--help | package-table --help]
  -> usage on stdout, exit 0

# validation
user -> kool go coverage [bad/missing args]
  -> stderr error, non-zero; no table

# package-table (L2 HandleWith)
user -> package-table [flags] <coverage.out>
  -> parse coverprofile -> filter/skip -> markdown|json on stdout
```

## Preconditions

- Module root is `d.DOCTEST_ROOT/../../..` (this tree lives at `tests/go/coverage/`).
- Product package `github.com/xhd2015/kool/tools/go/coverage` exports `Handle` /
  `HandleWith` with injectable `Stdout`/`Stderr` (see root `DOCTEST.md` DSN).
  Until implemented, suite is **RED**.
- `tools/go/go_tools.go` must eventually route `case "coverage"` (implementer);
  doctests call `coverage.HandleWith` directly and do not require a built binary.
- Fixtures use absolute paths under `req.WorkingDir` (from `d.DOCTEST_CASE`);
  no `os.Chdir` / `t.Chdir` / process env mutation.
- Classic coverprofile line form:
  `module/path/file.go:1.1,2.2 numStmts count`
  with `count > 0` counting all `numStmts` as covered (Python reference).

## Steps

1. Root `Setup` creates an isolated `WorkingDir` under `d.DOCTEST_CASE`.
2. Grouping/leaf `Setup` writes go.mod / coverage.out and sets flags/args.
3. `Run` calls `coverage.HandleWith` with capture buffers.

## Context

- Default module path in fixtures: `example.com/mod`.
- Default skips: prefixes `script/`, `cmd/`; contains `/legacy_`.
- Markdown headers always: `| Coverage | Package |` / `|----------|---------|`.
- Empty match: `warning:` on stderr + header-only table, exit 0.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = filepath.Join(d.DOCTEST_CASE, "work")
	}
	if err := os.MkdirAll(req.WorkingDir, 0755); err != nil {
		return err
	}
	return nil
}
```
