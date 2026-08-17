# Scenario

**Feature**: tasks.json parse validation

```
invalid JSONC / broken file -> Error on list or any command
```

## Steps

1. Leaves write invalid fixtures; typically invoke list.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// default exercise via list
	req.Subcommand = "list"
	return nil
}
```
