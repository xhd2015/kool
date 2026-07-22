# Scenario

**Feature**: dry-run errors on unknown ${…} variables

```
run "Bad Var" --dry-run with ${unknownToken}
  -> Error exit ≠ 0
```

## Steps

1. unresolvedVarJSONC; Query=`Bad Var`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, unresolvedVarJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Bad Var"
	return nil
}
```
