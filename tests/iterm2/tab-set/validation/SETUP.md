# Scenario

**Feature**: tab-set config validation rejects bad JSON schemas

```
invalid <name>.json -> show|run|list-load path -> Error exit ≠ 0
```

## Steps

1. Leaves write invalid fixtures and invoke show (or run --dry-run) to force load.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	// Default exercise path: show <name> after writing bad file.
	if req.Subcommand == "" {
		req.Subcommand = "show"
	}
	return nil
}
```
