# Scenario

**Feature**: default serve enables full teardown on Stop

```
kool cloudflare serve --domain a.example.com --url http://127.0.0.1:9
  -> StartSession Teardown=true
```

## Steps

1. Serve with domain+url only (no `--no-teardown`).

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
	req.NoWaitReady = true
	return nil
}
```
