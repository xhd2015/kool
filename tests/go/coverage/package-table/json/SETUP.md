# Scenario

**Feature**: --json emits stable sorted coverage rows

```
same two-package fixture as basic-sorted
user -> package-table --module example.com/mod --json <profile>
  -> JSON array [{"coverage":0,"package":"internal/run"},{"coverage":100,"package":"cli"}]
     (numeric coverage may be 0 / 0.0 / 100 / 100.0)
```

## Steps

1. Same profile as basic-sorted.
2. Set JSON true.

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
	req.JSON = true

	profile := filepath.Join(req.WorkingDir, "coverage.out")
	writeCoverProfile(t, profile, ""+
		"example.com/mod/cli/main.go:1.1,2.2 10 5\n"+
		"example.com/mod/internal/run/run.go:1.1,2.2 4 0\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	return nil
}
```
