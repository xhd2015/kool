# Scenario

**Feature**: --dir discovers module path from go.mod; foreign modules omitted

```
WorkingDir/go.mod: module example.com/mod
profile:
  example.com/mod/cli/a.go covered
  other.com/lib/b.go covered  (foreign)

user -> package-table --dir <WorkingDir> <profile>
  -> only `cli` row (module from go.mod; no --module flag)
```

## Steps

1. Write go.mod at WorkingDir.
2. Write profile with in-module + foreign lines.
3. Pass absolute --dir and profile; do not set --module.

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
		"example.com/mod/cli/a.go:1.1,2.2 4 4\n"+
		"other.com/lib/b.go:1.1,2.2 10 10\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	req.Dir = req.WorkingDir
	req.DirSet = true
	return nil
}
```
