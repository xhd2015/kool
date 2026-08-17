# Scenario

**Feature**: vscode tasks find — case-insensitive substring on label

```
find <query> -> matching task rows | error if zero
```

## Steps

1. Subcommand `find`.
2. Leaves set Query and fixture.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "find"
	return nil
}
```
