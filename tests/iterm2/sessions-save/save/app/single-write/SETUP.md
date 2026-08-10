# Scenario

**Feature**: single-app save always emits canonical `"app"` when known (system)

```
Caller
  -> sessions save --file app-single.json
  -> critical fixture (single AS source; preflight → /Applications/iTerm.app)
  <- FileJSON windows include "app": "/Applications/iTerm.app"; Saved; exit 0
```

## Steps

1. UseCriticalFixture; FixtureApp=system; FilePath=app-single.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseCriticalFixture = true
	req.FixtureApp = fixtureAppSystem
	req.FilePath = "app-single.json"
	return nil
}
```
