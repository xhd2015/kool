# Scenario

**Feature**: two packages with different coverage sort ascending in markdown

```
# profile under example.com/mod
cli/main.go: 10 stmts count>0  -> 100.0%
internal/run/run.go: 4 stmts count=0 -> 0.0%

user -> package-table --module example.com/mod <profile>
  -> markdown rows: internal/run 0.0% then cli 100.0%
```

## Steps

1. Write coverprofile with two packages (no go.mod needed when --module set).
2. Pass absolute profile path and --module.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "package-table"
	req.Module = "example.com/mod"
	req.ModuleSet = true

	profile := filepath.Join(req.WorkingDir, "coverage.out")
	writeCoverProfile(t, profile, ""+
		"example.com/mod/cli/main.go:1.1,2.2 10 5\n"+
		"example.com/mod/internal/run/run.go:1.1,2.2 4 0\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	return nil
}
```
