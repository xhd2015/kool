# Scenario

**Feature**: tab-set help subcommand

```
kool iterm2 tab-set --help -> usage mentioning list/run and config
```

## Steps

1. Leaves set Help or Subcommand for help path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	// Grouping: help leaves set Help=true (or equivalent).
	req.Subcommand = ""
	return nil
}
```

