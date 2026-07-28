# Scenario

**Feature**: tab-set --help prints usage and config path

```
tab-set --help -> exit 0; stdout has tab-set, list, run, config path
```

## Steps

1. Request Help on tab-set.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.Help = true
	return nil
}
```
