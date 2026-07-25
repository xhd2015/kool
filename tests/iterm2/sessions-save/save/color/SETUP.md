# Scenario

**Feature**: save dry-run color force flags (`--color` / `--no-color`)

```
Caller
  -> sessions save --dry-run --color | --no-color | both
  -> color resolver
  <- ANSI plan tokens | monochrome | conflict error
```

## Steps

1. ModeSave; DryRun; UseCriticalFixture (shared by color leaves).

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
	req.FilePath = "color-plan.json"
	return nil
}
```
