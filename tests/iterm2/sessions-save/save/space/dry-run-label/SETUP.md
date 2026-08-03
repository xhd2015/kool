# Scenario

**Feature**: save --dry-run plan shows `space N (Desktop N+1)`; never iterm_window_id

```
Caller
  -> sessions save --dry-run --file space-plan.json
  -> critical fixture (Space recording on)
  <- window block includes space N (Desktop N+1); no iterm_window_id text; no file
```

## Steps

1. DryRun; FilePath=space-plan.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "space-plan.json"
	return nil
}
```
