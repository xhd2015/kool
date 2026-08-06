# Scenario

**Feature**: package-table parses coverprofiles and renders filtered package rows

```
# default happy path
fixture go.mod + coverage.out under WorkingDir
  -> package-table --dir <abs> --module example.com/mod <profile>
  -> markdown/json rows sorted by coverage then package
```

## Steps

1. Grouping sets Subcommand `package-table`.
2. Leaves write go.mod + coverage.out and set filter/format flags.
3. Always pass absolute `--dir` and profile paths (parallel-safe).

## Context

- Fixture module: `example.com/mod`.
- Cover semantics: `numStmts` total; if `count > 0` then covered += `numStmts`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "package-table"
	// Leaves write fixtures under WorkingDir and set absolute --dir / profile / filters.
	return nil
}
```
