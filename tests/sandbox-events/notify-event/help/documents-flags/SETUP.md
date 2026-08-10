# Scenario

**Feature**: notify-event --help documents --type and --path

```
kool sandbox notify-event --help
  -> exit 0; stdout mentions --type and --path; trailing newline
```

## Steps

1. Inherit HelpNotifyEvent from parent.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpNotifyEvent = true
	req.RunNotifyEvent = false
	return nil
}
```
