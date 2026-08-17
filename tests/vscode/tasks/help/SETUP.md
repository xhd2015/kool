# Scenario

**Feature**: kool vscode tasks help

```
kool vscode tasks -h|--help -> usage on stdout exit 0
```

## Steps

1. Leaves set Help=true.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: help leaves set Help=true; clear subcommand.
	req.Subcommand = ""
	return nil
}
```
