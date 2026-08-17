# Scenario

**Feature**: serve without --url is invalid

```
kool cloudflare serve --domain a.example.com
  -> non-zero; message mentions url; no StartSession
```

## Steps

1. serve with Domain only.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "serve"
	req.DomainSet = true
	req.Domain = "a.example.com"
	req.URLSet = false
	return nil
}
```
