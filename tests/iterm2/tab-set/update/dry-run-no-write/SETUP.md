# Scenario

**Feature**: `--dry-run` prints plan and does not write

```
update bots --tab-id b --no-submit --dry-run
  -> plan/summary on stdout; bots.json still without no_submit on b
```

## Steps

1. Write bots fixture.
2. SetName=bots; TabID=b; UpdateNoSubmit=true; DryRun=true.

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
	req.DryRun = true
	return nil
}
```
