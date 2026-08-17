# Scenario

**Feature**: root --help lists serve and tunnel flags

```
kool cloudflare --help
  -> exit 0; stdout mentions serve, --domain, --url, --tunnel
```

## Steps

1. HelpAtRoot=true.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpAtRoot = true
	req.Subcommand = ""
	return nil
}
```
