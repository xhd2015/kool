# Scenario

**Feature**: adhoc --load-devbox at run merges secondary files into the session

```
# primary has from-primary.txt; secondary sealed has from-secondary.txt
kool sandbox build -o sandbox.bin --file from-primary.txt=from-primary.txt
# secondary: SecondaryPacks → load-secondary.bin
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin --load-devbox ABS_SECONDARY -- \
  sh -c 'cat from-secondary.txt'
  -> exit 0; stdout has notice + secondary content
```

## Steps

1. Pack primary with `from-primary.txt`.
2. SecondaryPack builds `load-secondary.bin` with `from-secondary.txt`.
3. SealedLoadDevbox = abs path of secondary; guest cats secondary file.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "from-primary.txt", "content-from-primary\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "from-secondary.txt", "content-from-secondary\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"from-primary.txt=from-primary.txt"}

	secOut := "load-secondary.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"from-secondary.txt=from-secondary.txt"},
	}}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedArgs = []string{"sh", "-c", "cat from-secondary.txt"}
	return nil
}
```
