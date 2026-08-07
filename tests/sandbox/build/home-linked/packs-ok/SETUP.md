# Scenario

**Feature**: --home-linked with a non-empty pack produces a sealed binary

```
kool sandbox build -o sandbox.bin --home-linked --file local.txt=app.txt --env MARKER=1
  -> exit 0; binary exists size>0
```

## Steps

1. Write a local source file; pass `--file` + `--env` (HomeLinked from parent).
2. No sealed run — pack-bit honor is proven by run/home-linked leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "local.txt", "home-linked pack body\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"local.txt=app.txt"}
	req.ExtraEnv = []string{"MARKER=1"}
	return nil
}
```
