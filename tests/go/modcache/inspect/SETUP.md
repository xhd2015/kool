# Scenario

**Feature**: inspect reports a fake $GOMODCACHE

```
user -> inspect --modcache <dir>
  -> summary on stdout; exit 0; no deletes
```

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "inspect"
	if req.ModCache == "" {
		req.ModCache = filepath.Join(req.WorkingDir, "mod")
	}
	req.ModCacheSet = true
	if err := os.MkdirAll(req.ModCache, 0755); err != nil {
		return err
	}
	return nil
}
```
