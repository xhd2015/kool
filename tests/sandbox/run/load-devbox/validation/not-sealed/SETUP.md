# Scenario

**Feature**: absolute path that is not a sealed sandbox binary fails hard

```
# plain text file at ABS
./primary --load-devbox ABS -- sh -c 'true'
  -> non-zero; stderr Error: (not sealed / unseal / invalid)
```

## Steps

1. Write a plain non-sealed file at abs path under WorkingDir.
2. SealedLoadDevbox points at that file; primary packs MARKER env.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	plain, err := writeLocalFile(t, req.WorkingDir, "not-a-sealed-binary.bin", "this is not a sealed sandbox\n")
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(plain)
	if err != nil {
		return err
	}
	req.ExtraEnv = []string{"MARKER=1"}
	req.SealedLoadDevbox = []string{abs}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
