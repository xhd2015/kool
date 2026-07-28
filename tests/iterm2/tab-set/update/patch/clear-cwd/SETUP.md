# Scenario

**Feature**: `--clear-cwd` empties cwd on an existing tab

```
bots tab b has cwd /tmp
  -> update bots --tab-id b --clear-cwd
  -> tab b cwd empty/omitted; command unchanged
```

## Steps

1. Write bots fixture (tab b has `"cwd": "/tmp"`).
2. SetName=bots; TabID=b; ClearCwd=true.

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
	req.ClearCwd = true
	return nil
}
```
