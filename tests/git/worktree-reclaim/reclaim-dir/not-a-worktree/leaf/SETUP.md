# Scenario

**Feature**: reclaim rejects main repo path

```
# main checkout is not a reclaim target
user -> kool git worktree reclaim <main-repo> -> error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.Path != req.MainRepo {
		t.Fatalf("expected Path to be main repo, got %q", req.Path)
	}
	return nil
}
```