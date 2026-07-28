# Scenario

**Feature**: update `--create` appends a new tab (requires `--command`)

```
existing bots.json
  -> update bots --tab-id c --create --command "echo c"
  -> tabs a,b,c present; c.command = echo c
```

## Steps

1. Leaves write bots (or other) fixture.
2. Set Create and related flags per leaf.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Create = true
	return nil
}
```
