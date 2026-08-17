# Scenario

**Feature**: reclaim rejects missing path

```
# filesystem path not found
user -> kool git worktree reclaim <missing> -> error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if pathExists(t, req.Path) {
		t.Fatalf("expected missing path, but exists: %s", req.Path)
	}
	return nil
}
```