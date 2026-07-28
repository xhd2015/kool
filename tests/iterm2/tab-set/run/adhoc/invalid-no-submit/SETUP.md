# Scenario

**Feature**: invalid no_submit prop value is an error (not silent ignore)

```
run scratch --tab "[no_submit=maybe] echo x" --dry-run
  -> Error exit ≠ 0; message mentions no_submit / invalid
```

## Steps

1. --tab with no_submit=maybe (not true/false/1/0/yes/no).
2. DryRun=true (parse fails before any iTerm).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "scratch"
	req.DryRun = true
	req.Tabs = []string{`[no_submit=maybe] echo x`}
	return nil
}
```
