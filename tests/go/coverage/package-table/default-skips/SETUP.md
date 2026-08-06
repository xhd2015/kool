# Scenario

**Feature**: default skip rules omit script/, cmd/, and /legacy_ packages

```
# keep: pkg/keep (50%)
# skip: script/x, cmd/y, pkg/legacy_z
user -> package-table --module example.com/mod <profile>
  -> only `pkg/keep` row
```

## Steps

1. Profile with keep + three skipped paths under module.
2. Defaults only (no custom skip flags).

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
		"example.com/mod/pkg/keep/k.go:1.1,2.2 2 1\n"+
		"example.com/mod/script/gen/g.go:1.1,2.2 8 8\n"+
		"example.com/mod/cmd/tool/main.go:1.1,2.2 8 8\n"+
		"example.com/mod/pkg/legacy_old/x.go:1.1,2.2 8 8\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	return nil
}
```
