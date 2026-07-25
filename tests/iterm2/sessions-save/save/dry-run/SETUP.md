# Scenario

**Feature**: save --dry-run with grok + mark fixture; no file written

```
Caller
  -> sessions save --dry-run --file plan.json
  -> fixture: one window grok + mark + idle
  <- Would save + session ids; plan.json not created
  # auto color: pipe/non-TTY → monochrome OK
```

## Steps

1. ModeSave; DryRun; UseCriticalFixture; FilePath=plan.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSave
	req.DryRun = true
	req.UseCriticalFixture = true
	req.FilePath = "plan.json"
	return nil
}
```
