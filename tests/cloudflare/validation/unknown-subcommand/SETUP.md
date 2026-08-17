# Scenario

**Feature**: unknown subcommand is rejected

```
kool cloudflare nosuch
  -> non-zero; stderr indicates unknown / unrecognized command
```

## Steps

1. Subcommand = nosuch.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "nosuch"
	return nil
}
```
