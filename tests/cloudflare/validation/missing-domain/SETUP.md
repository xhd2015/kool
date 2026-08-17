# Scenario

**Feature**: serve without --domain is invalid

```
kool cloudflare serve --url http://127.0.0.1:9
  -> non-zero; message mentions domain; no StartSession
```

## Steps

1. serve with URL only.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "serve"
	req.URLSet = true
	req.URL = "http://127.0.0.1:9"
	req.DomainSet = false
	return nil
}
```
