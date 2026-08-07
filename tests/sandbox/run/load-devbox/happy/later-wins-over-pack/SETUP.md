# Scenario

**Feature**: same relative path in primary and load — load content wins (later layer)

```
# primary shared.txt=P; secondary shared.txt=L
./primary --load-devbox ABS -- sh -c 'cat shared.txt'
  -> exit 0; stdout == L (load content, not primary)
```

## Steps

1. Primary packs `shared.txt` with primary content.
2. Secondary packs same relpath with load content.
3. Guest cats `shared.txt` — must see load content only.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := writeLocalFile(t, req.WorkingDir, "primary-shared.txt", "content-P-from-primary\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "load-shared.txt", "content-L-from-load\n"); err != nil {
		return err
	}
	secOut := "load-later-wins.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"load-shared.txt=shared.txt"},
	}}
	req.ExtraFiles = []string{"primary-shared.txt=shared.txt"}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedArgs = []string{"sh", "-c", "cat shared.txt"}
	return nil
}
```
