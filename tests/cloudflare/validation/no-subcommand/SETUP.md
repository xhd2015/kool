# Scenario

**Feature**: bare root with no subcommand is invalid

```
kool cloudflare
  -> non-zero; stderr suggests subcommands / help / usage
```

## Steps

1. Empty Subcommand; no help flags.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = ""
	return nil
}
```
