# Scenario

**Feature**: inspect with a missing --modcache directory fails

```
user -> inspect --modcache /abs/missing
  -> exit 1; Error: on stderr
```

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "inspect"
	req.ModCache = filepath.Join(req.WorkingDir, "no-such-modcache")
	req.ModCacheSet = true
	return nil
}
```
