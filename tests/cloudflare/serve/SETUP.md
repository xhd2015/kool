# Scenario

**Feature**: serve lifecycle with injected StartSession / WaitSignal

```
# happy
user -> serve --domain HOST --url URL [--tunnel NAME]
  -> StartSession(Domain, LocalURL, TunnelName)
  -> print public URL → WaitSignal → Session.Stop → exit 0

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
