# Scenario

**Feature**: tab-set help documents --tab, --save, --force, and no_submit prop

```
tab-set --help -> mentions --tab, --save, --force; ad-hoc props include no_submit
```

## Steps

1. Help=true (same as show-usage; stronger content asserts).

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
