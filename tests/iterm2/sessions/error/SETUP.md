# Scenario

**Feature**: sessions snapshot error paths (no successful stream)

```
# iTerm not running
Capture / snapshot with ITermRunning=false
  -> non-zero; stderr mentions iTerm2 / not running

# format conflict
sessions snapshot --json --html
  -> non-zero; mutually exclusive; no capture required
```

## Steps

1. Mode=snapshot-cli by default for error leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSnapshotCLI
	return nil
}
```
