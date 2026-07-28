# Scenario

**Feature**: update on unknown set name errors

```
update no-such-set --tab-id a --no-submit
  -> not found / missing set error; exit ≠ 0
```

## Steps

1. Empty ConfigDir (no fixture).
2. SetName=no-such-set; TabID=a; UpdateNoSubmit=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "no-such-set"
	req.TabID = "a"
	req.UpdateNoSubmit = true
	return nil
}
```
