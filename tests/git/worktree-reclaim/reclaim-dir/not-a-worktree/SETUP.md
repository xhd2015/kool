# Scenario

**Feature**: main repo path is not a linked worktree

```
# path points at main checkout, not a linked worktree
reclaim handler -> isLinked(path)=false -> error
```

## Steps

1. Create a main git repository without using it as a linked worktree target

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	req.MainRepo = mainRepo
	req.Path = mainRepo
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```