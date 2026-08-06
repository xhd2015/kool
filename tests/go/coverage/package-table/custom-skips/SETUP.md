# Scenario

**Feature**: --skip-prefix and --skip-contains replace default skip lists

```
# With custom skips only "tmp/" prefix and "/x_drop" contains:
#   keep: script/gen (default skip disabled by override)
#   keep: pkg/keep
#   skip: tmp/x
#   skip: pkg/x_drop/z

user -> package-table --module example.com/mod \
         --skip-prefix tmp/ --skip-contains /x_drop <profile>
  -> rows for pkg/keep and script/gen only (sorted)
```

## Steps

1. Profile includes default-skip path (script/) that must be kept under custom flags.
2. Set custom --skip-prefix and --skip-contains (replacing defaults).

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
	req.SkipPrefix = "tmp/"
	req.SkipPrefixSet = true
	req.SkipContains = "/x_drop"
	req.SkipContainsSet = true

	profile := filepath.Join(req.WorkingDir, "coverage.out")
	writeCoverProfile(t, profile, ""+
		"example.com/mod/pkg/keep/k.go:1.1,2.2 2 0\n"+
		"example.com/mod/script/gen/g.go:1.1,2.2 2 2\n"+
		"example.com/mod/tmp/x/t.go:1.1,2.2 8 8\n"+
		"example.com/mod/pkg/x_drop/z.go:1.1,2.2 8 8\n")
	req.ProfilePath = profile
	req.ProfileSet = true
	return nil
}
```
