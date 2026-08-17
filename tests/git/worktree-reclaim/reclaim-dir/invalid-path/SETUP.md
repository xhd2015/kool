# Scenario

**Feature**: nonexistent path returns error

```
# path does not exist on filesystem
reclaim handler -> stat path -> error
```

## Steps

1. Point reclaim at a path that does not exist

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t)
	req.MainRepo = mainRepo
	req.Path = filepath.Join(filepath.Dir(mainRepo), "does-not-exist")
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```