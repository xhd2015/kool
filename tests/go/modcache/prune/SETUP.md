# Scenario

**Feature**: prune deletes legacy versions of a module

```
foo@v1.0.0 + foo@v1.2.0 + two toolchains
prune [--dry-run] --modcache <dir>
  -> dry-run keeps files; apply removes v1.0.0 and keeps v1.2.0 + toolchain
```

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "prune"
	req.ModCache = filepath.Join(req.WorkingDir, "mod")
	req.ModCacheSet = true
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.0.0", "old\n")
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.2.0", "new\n")
	writeZip(t, req.ModCache, "example.com/foo", "v1.0.0", "OLDZIP")
	writeZip(t, req.ModCache, "example.com/foo", "v1.2.0", "NEWZIP")
	writeExtracted(t, req.ModCache, "golang.org/toolchain", "v0.0.1-go1.21.0.darwin-arm64", "t1")
	writeExtracted(t, req.ModCache, "golang.org/toolchain", "v0.0.1-go1.22.0.darwin-arm64", "t2")
	return nil
}
```
