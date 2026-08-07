# Scenario

**Feature**: relative --runtime-load-devbox path is rejected at pack time

```
kool sandbox build -o sandbox.bin --file a.txt=a.txt \
  --runtime-load-devbox relative/path
  -> non-zero; stderr Error: (absolute / relative / runtime-load-devbox)
```

## Steps

1. Non-empty pack via `--file` so failure is path form, not empty-pack.
2. Pass a relative `--runtime-load-devbox` path (as-is; not Abs-resolved).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "a.txt", "relative reject body\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"a.txt=a.txt"}
	req.RuntimeLoadDevbox = []string{"relative/path"}
	req.AfterBuildInspect = false
	return nil
}
```
