# Scenario

**Feature**: vscode tasks show — exact or unique CI substring

```
show <label> -> task details | missing/ambiguous error
```

## Steps

1. Subcommand `show`.
2. Leaves set Query and fixture.

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
