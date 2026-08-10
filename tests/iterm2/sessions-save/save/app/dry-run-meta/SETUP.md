# Scenario

**Feature**: save --dry-run shows gray `app  …` meta when app non-empty

```
Caller
  -> sessions save --dry-run --file app-plan.json
  -> critical fixture (system app known)
  <- window block includes "app" meta line; no leading blank; no file; exit 0
```

## Steps

1. DryRun; UseCriticalFixture; FixtureApp=system; FilePath=app-plan.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.UseCriticalFixture = true
	req.FixtureApp = fixtureAppSystem
	req.FilePath = "app-plan.json"
	return nil
}
```
