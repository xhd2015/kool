# Scenario

**Feature**: --tab props block allows arbitrary spaces around [ ] and key=value

```
run scratch --tab "  [ id = a ]  echo hi" --dry-run
  -> id=a command=echo hi; exit 0
```

## Steps

1. Single --tab with leading/trailing spaces and spaces inside props.
2. DryRun=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "scratch"
	req.DryRun = true
	req.Tabs = []string{"  [ id = a ]  echo hi"}
	return nil
}
```
