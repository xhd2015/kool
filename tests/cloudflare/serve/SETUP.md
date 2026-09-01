# Scenario

**Feature**: serve lifecycle with injected StartSession / WaitSignal

```
# happy
user -> serve --domain HOST --url URL [--tunnel NAME]
  -> StartSession(Domain, LocalURL, TunnelName, Teardown=true)
  -> print public URL → WaitReady → WaitSignal → Session.Stop → exit 0

# wait timeout
WaitReady ErrReadyTimeout → warning on stderr; still WaitSignal → Stop → exit 0

# no-wait-ready
--no-wait-ready → skip WaitReady

# no-teardown
--no-teardown → StartSession Teardown=false

# start failure
StartSession error → non-zero; stderr surfaces error
```

## Steps

1. AllowStart=true so inject may succeed or return StartSessionErr.
2. Leaves set Domain/URL/Tunnel and optional StartSessionErr.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "serve"
	req.AllowStart = true
	req.HelpAtRoot = false
	req.HelpServe = false
	return nil
}
```
