# Scenario

**Feature**: --module overrides go.mod and filters to explicit prefix

```
# go.mod says example.com/mod but --module example.com/other
profile:
  example.com/other/svc/s.go covered
  example.com/mod/cli/c.go covered

user -> package-table --dir <dir> --module example.com/other <profile>
  -> only `svc` (explicit --module wins over go.mod)
```

## Steps

1. Write go.mod for example.com/mod (should be ignored when --module set).
2. Profile with both prefixes; set --module example.com/other.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "package-table"
	writeGoMod(t, req.WorkingDir, "example.com/mod")

	profile := filepath.Join(req.WorkingDir, "coverage.out")
	writeCoverProfile(t, profile, ""+
		"example.com/other/svc/s.go:1.1,2.2 2 2\n"+
		"example.com/mod/cli/c.go:1.1,2.2 8 8\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	req.Dir = req.WorkingDir
	req.DirSet = true
	req.Module = "example.com/other"
	req.ModuleSet = true
	return nil
}
```
