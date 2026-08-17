# Scenario

**Feature**: run unique CI substring match with --dry-run

```
run "serv" --dry-run matches Serve only -> exit 0
```

## Steps

1. Multi-task; Query=`serv` (substring of Serve).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "serv"
	return nil
}
```
