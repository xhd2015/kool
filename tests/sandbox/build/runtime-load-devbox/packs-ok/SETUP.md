# Scenario

**Feature**: absolute --runtime-load-devbox paths seal; inspect lists them

```
# two abs paths need not exist on disk at pack time
kool sandbox build -o sandbox.bin --file a.txt=a.txt \
  --runtime-load-devbox /abs/.../box-a \
  --runtime-load-devbox /abs/.../box-b
  -> exit 0; binary exists
kool sandbox inspect sandbox.bin
  -> exit 0; lists both absolute paths under runtime-load-devbox
```

## Steps

1. Write a local `--file` source so the pack is non-empty (load paths alone do
   not fill pack).
2. Build two absolute path strings under WorkingDir (paths need not exist).
3. Enable AfterBuildInspect to verify sealed list.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "a.txt", "runtime-load-devbox pack body\n"); err != nil {
		return err
	}
	// Absolute remote-style paths under WorkingDir; do not create the files —
	// pack seals strings only (no FS existence check).
	boxA := filepath.Join(req.WorkingDir, "remote-boxes", "box-a")
	boxB := filepath.Join(req.WorkingDir, "remote-boxes", "box-b")
	absA, err := filepath.Abs(boxA)
	if err != nil {
		return err
	}
	absB, err := filepath.Abs(boxB)
	if err != nil {
		return err
	}
	req.ExtraFiles = []string{"a.txt=a.txt"}
	req.RuntimeLoadDevbox = []string{absA, absB}
	req.AfterBuildInspect = true
	return nil
}
```
