# Scenario

**Feature**: same env key in primary and load → hard Error (even different values)

```
# primary FOO=1; secondary FOO=2
./primary --load-devbox ABS -- sh -c 'true'
  -> non-zero; stderr Error: + FOO + incompatible / both sources
```

## Steps

1. Primary `--env FOO=1`; secondary `--env FOO=2`.
2. Guest is trivial — must fail during env merge before success.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	secOut := "load-env-conflict-primary.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:   secOut,
		ExtraEnv: []string{"FOO=2"},
	}}
	req.ExtraEnv = []string{"FOO=1"}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
