# Scenario

**Feature**: empty match after filters → warning + header-only table, exit 0

```
# profile only has skipped or foreign lines under default filter
user -> package-table --module example.com/mod <profile>
  -> exit 0
  -> stdout header-only markdown table
  -> stderr warning: (no packages matched)
```

## Steps

1. Profile with only script/ and foreign module lines (nothing kept).
2. Default skips + module filter.

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
		"example.com/mod/script/gen/g.go:1.1,2.2 8 8\n"+
		"other.com/lib/b.go:1.1,2.2 10 10\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	return nil
}
```
