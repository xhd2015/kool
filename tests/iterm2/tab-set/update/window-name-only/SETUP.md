# Scenario

**Feature**: `--window-name` alone (no `--tab-id`) patches set-level window_name

```
update bots --window-name new-bots-win
  -> window_name becomes new-bots-win; tabs unchanged
```

## Steps

1. Write bots fixture (window_name local-bots).
2. SetName=bots; WindowName=new-bots-win; TabID empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.WindowName = "new-bots-win"
	// TabID intentionally empty — set-level patch only.
	return nil
}
```
