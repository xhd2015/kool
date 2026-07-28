# Scenario

**Feature**: --save persists no_submit=true on version-1 JSON; never runs iTerm

```
run mysave --tab "[id=x] echo x" --tab "[id=y,no_submit=true] echo y" --save --force
  -> mysave.json version 1; tab y has "no_submit": true; tab x omits or false
```

## Steps

1. Empty ConfigDir.
2. Two ad-hoc tabs; Save=true; Force=true; WindowName optional.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "mysave"
	req.Save = true
	req.Force = true
	req.WindowName = "win-save"
	req.Tabs = []string{
		`[id=x] echo x`,
		`[id=y,no_submit=true] echo y`,
	}
	return nil
}
```
