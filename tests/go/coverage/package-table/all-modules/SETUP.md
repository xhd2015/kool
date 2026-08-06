# Scenario

**Feature**: --all disables module filter so foreign packages appear

```
profile:
  example.com/mod/cli/a.go 100%
  other.com/lib/b.go 0%

user -> package-table --all --module example.com/mod <profile>
  -> both packages appear (sorted); --all wins over module filter
```

## Steps

1. Write profile with two modules.
2. Pass --all (and optional --module that would otherwise exclude foreign).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "package-table"
	req.All = true
	// Module still set to show --all overrides filter even if module is present.
	req.Module = "example.com/mod"
	req.ModuleSet = true

	profile := filepath.Join(req.WorkingDir, "coverage.out")
	writeCoverProfile(t, profile, ""+
		"example.com/mod/cli/a.go:1.1,2.2 4 4\n"+
		"other.com/lib/b.go:1.1,2.2 4 0\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	return nil
}
```
