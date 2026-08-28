# Scenario

**Feature**: WaitReady timeout warns on stderr and keeps serving

```
kool cloudflare serve --domain a.example.com --url http://127.0.0.1:9 --ready-timeout 5s
  -> StartSession
  -> WaitReady returns ErrReadyTimeout
  -> stderr warning; WaitSignal; Stop; exit 0
```

## Steps

1. Serve with short ready-timeout; WaitReadyMode=timeout.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DomainSet = true
	req.Domain = "a.example.com"
	req.URLSet = true
	req.URL = "http://127.0.0.1:9"
	req.ReadyTimeoutSet = true
	req.ReadyTimeout = "5s"
	req.WaitReadyMode = "timeout"
	return nil
}
```
