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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t)
	req.MainRepo = mainRepo
	req.Path = mainRepo
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```