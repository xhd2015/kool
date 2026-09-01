# Scenario

**Feature**: `--no-teardown` keeps DNS/tunnel on Stop

```
kool cloudflare serve --domain a.example.com --url http://127.0.0.1:9 --no-teardown
  -> StartSession Teardown=false
```

## Steps

1. Serve with `--no-teardown`.

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
	req.NoTeardown = true
	return nil
}
```
