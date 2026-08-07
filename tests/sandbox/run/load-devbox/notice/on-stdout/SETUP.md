# Scenario

**Feature**: successful load prints notice: loading devbox on stdout (no ANSI in capture)

```
./primary --load-devbox ABS -- sh -c 'true'
  -> exit 0; RunStdout contains "notice: loading devbox <abs>\n"; no ESC sequences
```

## Steps

1. Minimal primary + secondary so load succeeds.
2. Guest `true`; assert focuses on notice line format.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "p.txt", "p\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "s.txt", "s\n"); err != nil {
		return err
	}
	secOut := "load-notice.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"s.txt=s.txt"},
	}}
	req.ExtraFiles = []string{"p.txt=p.txt"}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
