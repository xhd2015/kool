# Scenario

**Feature**: update with no action flags errors “nothing to update”

```
update bots
  -> error nothing to update; file unchanged
```

## Steps

1. Write bots fixture.
2. SetName=bots only — no TabID, no window-name, no patch flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	// intentionally no action flags
	return nil
}
```
