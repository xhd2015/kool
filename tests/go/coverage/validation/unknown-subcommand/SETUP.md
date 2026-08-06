# Scenario

**Feature**: unknown coverage subcommand errors clearly

```
user -> HandleWith(["nosuch"])
  -> non-zero; stderr indicates unknown/unrecognized
```

## Steps

1. Subcommand `nosuch`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"nosuch"}
	return nil
}
```
