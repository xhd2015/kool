# Scenario

**Feature**: `--no-submit` sets `no_submit: true` on an existing tab

```
update bots --tab-id b --no-submit
  -> bots.json tab b has "no_submit": true; tab a unchanged
  -> stdout short summary (no_submit false -> true)
```

## Steps

1. Write bots fixture (tabs a,b; no no_submit set).
2. SetName=bots; TabID=b; UpdateNoSubmit=true.

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
	return nil
}
```
