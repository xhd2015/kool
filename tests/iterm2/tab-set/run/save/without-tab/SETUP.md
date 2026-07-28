# Scenario

**Feature**: --save without any --tab is an error

```
run bots --save -> Error; no write
```

## Steps

1. Optional bots fixture present (should not matter).
2. Save=true; Tabs empty; no DryRun required.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.Save = true
	// Tabs intentionally empty — config mode cannot --save.
	return nil
}
```
