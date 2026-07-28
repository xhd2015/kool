# Scenario

**Feature**: -n / --new-window with --save is an error

```
run scratch --tab "echo x" --save -n -> Error (save-only; window flags unused)
```

## Steps

1. Save + Tabs + NewWindow; no Force needed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "scratch"
	req.Save = true
	req.NewWindow = true
	req.Tabs = []string{"echo x"}
	return nil
}
```
