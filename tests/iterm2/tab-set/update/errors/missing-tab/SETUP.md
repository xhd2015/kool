# Scenario

**Feature**: patch missing tab id without `--create` errors

```
update bots --tab-id missing --no-submit
  -> tab not found (hint --create); file unchanged
```

## Steps

1. Write bots fixture.
2. SetName=bots; TabID=missing; UpdateNoSubmit=true; Create=false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.TabID = "missing"
	req.UpdateNoSubmit = true
	return nil
}
```
