# Scenario

**Feature**: ad-hoc --tab prop no_submit=true appears in --dry-run plan

```
run scratch --tab "[id=x,no_submit=true] echo staged" --dry-run
  -> exit 0; plan marks no_submit; no config file required
```

## Steps

1. Empty ConfigDir (no scratch.json).
2. One --tab with id=x and no_submit=true; DryRun=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "scratch"
	req.DryRun = true
	req.Tabs = []string{`[id=x,no_submit=true] echo staged`}
	return nil
}
```
