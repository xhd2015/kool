# Scenario

**Feature**: StopOnFirstArg — guest command after --load-devbox without leading --

```
# no double-dash before guest argv
./primary --load-devbox ABS sh -c 'echo ok-stop-on-first'
  -> exit 0; stdout contains ok-stop-on-first (and load notice)
```

## Steps

1. Secondary pack with a tiny file so load succeeds.
2. SealedDoubleDash=false; SealedArgs is guest only; SealedLoadDevbox set.
3. Harness argv: `--load-devbox ABS sh -c '…'` (no `--`).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "primary.txt", "p\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "sec.txt", "s\n"); err != nil {
		return err
	}
	secOut := "load-stop-first.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"sec.txt=sec.txt"},
	}}
	req.ExtraFiles = []string{"primary.txt=primary.txt"}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedDoubleDash = false
	req.SealedArgs = []string{"sh", "-c", "echo ok-stop-on-first"}
	return nil
}
```
