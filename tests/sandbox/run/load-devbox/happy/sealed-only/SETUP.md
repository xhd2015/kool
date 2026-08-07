# Scenario

**Feature**: sealed RuntimeLoadDevbox alone loads secondary at run (no CLI flag)

```
# secondary built first; primary seals its abs path at pack time
kool sandbox build -o sandbox.bin --file from-primary.txt=from-primary.txt \
  --runtime-load-devbox ABS_SECONDARY
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'cat from-sealed-load.txt'
  -> exit 0; notice + sealed-load file content
```

## Steps

1. SecondaryPack with `from-sealed-load.txt`.
2. Primary packs `from-primary.txt` and RuntimeLoadDevbox = abs secondary.
3. No SealedLoadDevbox (adhoc empty); guest cats sealed-load file.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "from-primary.txt", "primary-body\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "from-sealed-load.txt", "content-from-sealed-load\n"); err != nil {
		return err
	}
	secOut := "load-sealed-only.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"from-sealed-load.txt=from-sealed-load.txt"},
	}}
	req.ExtraFiles = []string{"from-primary.txt=from-primary.txt"}
	req.RuntimeLoadDevbox = []string{secAbs}
	req.SealedLoadDevbox = nil
	req.SealedArgs = []string{"sh", "-c", "cat from-sealed-load.txt"}
	return nil
}
```
