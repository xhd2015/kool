# Scenario

**Feature**: --no-tree keeps agent session id but omits FormatTree lines

```
sessions snapshot --no-color --no-tree
  + busyGrokResolve inject
  -> stdout contains agent session uuid + grok
  -> stdout must NOT contain ├── or └──
```

## Steps

1. NoEnrich=false; NoTree=true; default busyGrokResolve.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoEnrich = false
	req.NoTree = true
	return nil
}
```
