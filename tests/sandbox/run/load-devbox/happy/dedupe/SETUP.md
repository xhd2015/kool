# Scenario

**Feature**: same abs path in sealed RuntimeLoadDevbox and CLI loads once (first-seen)

```
# same ABS sealed + --load-devbox ABS
./primary --load-devbox ABS -- sh -c 'cat from-load.txt'
  -> exit 0; exactly one "notice: loading devbox ABS"
```

## Steps

1. One secondary with `from-load.txt`.
2. Primary RuntimeLoadDevbox and SealedLoadDevbox both set to that abs path.
3. Guest cats load file; assert single notice occurrence.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "primary.txt", "primary\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "from-load.txt", "content-dedupe-load\n"); err != nil {
		return err
	}
	secOut := "load-dedupe.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"from-load.txt=from-load.txt"},
	}}
	req.ExtraFiles = []string{"primary.txt=primary.txt"}
	req.RuntimeLoadDevbox = []string{secAbs}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedArgs = []string{"sh", "-c", "cat from-load.txt"}
	return nil
}
```
