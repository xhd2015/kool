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
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	req.MainRepo = mainRepo
	req.Path = filepath.Join(filepath.Dir(mainRepo), "does-not-exist")
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```