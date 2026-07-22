# Scenario

**Feature**: run unique CI substring match with --dry-run

```
run "serv" --dry-run matches Serve only -> exit 0
```

## Steps

1. Multi-task; Query=`serv` (substring of Serve).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "serv"
	return nil
}
```
