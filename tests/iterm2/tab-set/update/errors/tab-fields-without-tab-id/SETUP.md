# Scenario

**Feature**: tab field patch without `--tab-id` is an error

```
update bots --no-submit
  -> error requires --tab-id; file unchanged
```

## Steps

1. Write bots fixture.
2. UpdateNoSubmit=true; TabID empty; no WindowName.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.UpdateNoSubmit = true
	// TabID intentionally empty
	return nil
}
```
