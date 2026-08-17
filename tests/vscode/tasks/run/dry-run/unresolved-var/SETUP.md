# Scenario

**Feature**: dry-run errors on unknown ${…} variables

```
run "Bad Var" --dry-run with ${unknownToken}
  -> Error exit ≠ 0
```

## Steps

1. unresolvedVarJSONC; Query=`Bad Var`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, unresolvedVarJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Bad Var"
	return nil
}
```
