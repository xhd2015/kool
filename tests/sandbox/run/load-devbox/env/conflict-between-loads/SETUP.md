# Scenario

**Feature**: same env key from two load-devbox packs → hard Error

```
# primary MARKER=ok; load-a FOO=1; load-b FOO=2
./primary --load-devbox A --load-devbox B -- sh -c 'true'
  -> non-zero; stderr Error: + FOO
```

## Steps

1. Two secondaries both set FOO (different values); primary uses disjoint MARKER.
2. SealedLoadDevbox lists both abs paths in order.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	outA := "load-env-a.bin"
	outB := "load-env-b.bin"
	absA := filepath.Join(req.WorkingDir, outA)
	absB := filepath.Join(req.WorkingDir, outB)
	req.SecondaryPacks = []SecondaryPack{
		{Output: outA, ExtraEnv: []string{"FOO=1"}},
		{Output: outB, ExtraEnv: []string{"FOO=2"}},
	}
	req.ExtraEnv = []string{"MARKER=ok"}
	req.SealedLoadDevbox = []string{absA, absB}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
