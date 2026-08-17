# Scenario

**Feature**: serve --help documents serve flags

```
kool cloudflare serve --help
  -> exit 0; stdout mentions --domain, --url, --tunnel
```

## Steps

1. HelpServe=true.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpServe = true
	return nil
}
```
