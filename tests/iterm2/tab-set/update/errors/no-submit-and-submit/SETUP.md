# Scenario

**Feature**: `--no-submit` and `--submit` are mutually exclusive

```
update bots --tab-id b --no-submit --submit
  -> exclusive error; file unchanged
```

## Steps

1. Write bots fixture.
2. TabID=b; UpdateNoSubmit=true; UpdateSubmit=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.TabID = "b"
	req.UpdateNoSubmit = true
	req.UpdateSubmit = true
	return nil
}
```
