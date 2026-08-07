# Scenario

**Feature**: disjoint env keys from primary and load both visible to guest

```
# primary --env PRIM_KEY=1; secondary --env LOAD_KEY=2
./primary --load-devbox ABS -- sh -c 'printf %s/%s "$PRIM_KEY" "$LOAD_KEY"'
  -> exit 0; stdout == 1/2
```

## Steps

1. Primary packs PRIM_KEY=1 (and a tiny file so pack non-empty if needed — env alone OK).
2. Secondary packs LOAD_KEY=2.
3. Guest prints both values.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	secOut := "load-env-disjoint.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:   secOut,
		ExtraEnv: []string{"LOAD_KEY=2"},
	}}
	req.ExtraEnv = []string{"PRIM_KEY=1"}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedArgs = []string{"sh", "-c", `printf '%s/%s' "$PRIM_KEY" "$LOAD_KEY"`}
	return nil
}
```
