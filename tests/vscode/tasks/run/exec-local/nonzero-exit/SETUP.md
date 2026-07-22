# Scenario

**Feature**: local backend surfaces non-zero child exit

```
run "Fail Fast" --backend=local
  -> command false -> exit ≠ 0
```

## Steps

1. failLeafJSONC; Query=`Fail Fast`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, failLeafJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Fail Fast"
	return nil
}
```
