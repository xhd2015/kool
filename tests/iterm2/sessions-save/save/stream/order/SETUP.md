# Scenario

**Feature**: save dry-run emits W1 before last window ListTabs runs

```
sessions save --dry-run --no-color
  -> progressive stream (two critical windows)
  -> when ListTabsAndSessions(2) starts, stdout already contains "W1"
  -> final stdout also has W2 and Would save footer
```

## Steps

1. Default progressive dry-run stream (ObserveStreamOrder already set by parent).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	return nil
}
```
