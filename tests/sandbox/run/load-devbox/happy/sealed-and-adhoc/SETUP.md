# Scenario

**Feature**: sealed RuntimeLoadDevbox and adhoc --load-devbox both apply (distinct paths)

```
# sealed load A + CLI load B (distinct abs paths)
./primary --load-devbox ABS_B -- sh -c 'cat sealed-a.txt; cat adhoc-b.txt'
  -> exit 0; two notices (A then B); both file contents
```

## Steps

1. SecondaryPacks: load-a.bin (sealed-a.txt), load-b.bin (adhoc-b.txt).
2. Primary RuntimeLoadDevbox = A; SealedLoadDevbox = B.
3. Guest cats both files (order: sealed-a then adhoc-b on separate lines).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "primary.txt", "primary-body\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "sealed-a.txt", "content-sealed-a\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "adhoc-b.txt", "content-adhoc-b\n"); err != nil {
		return err
	}
	outA := "load-a.bin"
	outB := "load-b.bin"
	absA := filepath.Join(req.WorkingDir, outA)
	absB := filepath.Join(req.WorkingDir, outB)
	req.SecondaryPacks = []SecondaryPack{
		{Output: outA, ExtraFiles: []string{"sealed-a.txt=sealed-a.txt"}},
		{Output: outB, ExtraFiles: []string{"adhoc-b.txt=adhoc-b.txt"}},
	}
	req.ExtraFiles = []string{"primary.txt=primary.txt"}
	req.RuntimeLoadDevbox = []string{absA}
	req.SealedLoadDevbox = []string{absB}
	req.SealedArgs = []string{"sh", "-c", "cat sealed-a.txt; cat adhoc-b.txt"}
	return nil
}
```
