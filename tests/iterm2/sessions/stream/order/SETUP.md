# Scenario

**Feature**: default CLI emits W1 before last window ListTabs runs

```
sessions snapshot --no-color
  -> progressive stream
  -> when ListTabsAndSessions(2) starts, stdout already contains "W1"
  -> final stdout also has W2 and session footer
```

## Steps

1. Default stream (NoStream=false).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoStream = false
	return nil
}
```
