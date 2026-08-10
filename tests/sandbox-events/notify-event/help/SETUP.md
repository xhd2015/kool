# Scenario

**Feature**: notify-event help mode (no publish)

```
kool sandbox notify-event --help
  -> usage; exit 0
```

## Steps

1. HelpNotifyEvent=true; clear RunNotifyEvent.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.ProcessTimeout > 15*time.Second || req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 15 * time.Second
	}
	req.HelpNotifyEvent = true
	req.RunNotifyEvent = false
	req.LiveSession = false
	return nil
}
```
