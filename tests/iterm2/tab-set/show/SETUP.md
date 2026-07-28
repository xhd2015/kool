# Scenario

**Feature**: tab-set show prints one config

```
tab-set show <name> -> window_name + tabs id/command
```

## Steps

1. Subcommand `show`; leaves set SetName and fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.Subcommand = "show"
	return nil
}
```
