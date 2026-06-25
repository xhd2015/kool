# Scenario

**Feature**: reclaim rejects missing path

```
# filesystem path not found
user -> kool git worktree reclaim <missing> -> error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if pathExists(t, req.Path) {
		t.Fatalf("expected missing path, but exists: %s", req.Path)
	}
	return nil
}
```