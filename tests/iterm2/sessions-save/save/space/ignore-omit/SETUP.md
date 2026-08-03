# Scenario

**Feature**: save --ignore-macos-space omits both space and iterm_window_id; no resolve

```
Caller
  -> sessions save --ignore-macos-space --file ignore-space.json
  -> critical fixture; no SpaceIndexForWindow
  <- FileJSON has neither "space" nor "iterm_window_id" keys
```

## Steps

1. IgnoreMacOSSpace; FilePath=ignore-space.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.IgnoreMacOSSpace = true
	req.FilePath = "ignore-space.json"
	return nil
}
```
