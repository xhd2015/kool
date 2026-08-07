# Scenario

**Feature**: nested sealed RuntimeLoadDevbox — primary loads B; B seals A; guest sees A files

```
# A packs nested-a.txt; B packs nested-b.txt + --runtime-load-devbox A
# primary packs primary.txt + --runtime-load-devbox B
./primary -- sh -c 'cat nested-a.txt'
  -> exit 0; content from A (DFS load of B's sealed list)
```

## Steps

1. SecondaryPacks order: build A first, then B with RuntimeLoadDevbox=A.
2. Primary RuntimeLoadDevbox=B only (no adhoc).
3. Guest cats nested-a.txt from deepest sealed load.

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
	if _, err := writeLocalFile(t, req.WorkingDir, "nested-a.txt", "content-from-nested-a\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "nested-b.txt", "content-from-nested-b\n"); err != nil {
		return err
	}
	outA := "nested-a.bin"
	outB := "nested-b.bin"
	absA := filepath.Join(req.WorkingDir, outA)
	absB := filepath.Join(req.WorkingDir, outB)
	// Build A then B (B seals path to A).
	req.SecondaryPacks = []SecondaryPack{
		{Output: outA, ExtraFiles: []string{"nested-a.txt=nested-a.txt"}},
		{
			Output:            outB,
			ExtraFiles:        []string{"nested-b.txt=nested-b.txt"},
			RuntimeLoadDevbox: []string{absA},
		},
	}
	req.ExtraFiles = []string{"primary.txt=primary.txt"}
	req.RuntimeLoadDevbox = []string{absB}
	req.SealedLoadDevbox = nil
	req.SealedArgs = []string{"sh", "-c", "cat nested-a.txt"}
	return nil
}
```
