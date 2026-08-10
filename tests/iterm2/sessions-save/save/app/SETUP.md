# Scenario

**Feature**: sessions save records per-window canonical `app` (multi-app merge)

```
Caller
  -> sessions save [--dry-run] [--spaces] --file …
  -> preflight asApp + dual sources (home + /Applications) when running
  -> merge / dedupe iterm_window_id / renumber source_index
  <- FileJSON "app" / dry-run gray app meta / collapse warn
```

## Steps

1. ModeSave (shared by app save leaves). Leaves set fixture flags + DryRun/Spaces.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSave
	return nil
}
```
