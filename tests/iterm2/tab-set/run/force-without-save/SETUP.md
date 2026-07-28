# Scenario

**Feature**: --force without --save is an error

```
run scratch --tab "echo x" --force --dry-run
  -> Error (--force only valid with --save)
```

## Steps

1. Tabs set; Force=true; Save=false; DryRun to avoid iTerm if force were ignored.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "scratch"
	req.Force = true
	req.Save = false
	req.DryRun = true
	req.Tabs = []string{"echo x"}
	return nil
}
```
