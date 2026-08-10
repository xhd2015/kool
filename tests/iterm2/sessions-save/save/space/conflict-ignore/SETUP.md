# Scenario

**Feature**: `--spaces` cannot combine with `--ignore-macos-space`

```
Caller
  -> sessions save --spaces 0 --ignore-macos-space
  <- non-zero; Error about cannot be used together
```

## Steps

1. Spaces=0; IgnoreMacOSSpace; no fixture required.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseCriticalFixture = false
	req.Spaces = "0"
	req.IgnoreMacOSSpace = true
	req.FilePath = "conflict.json"
	return nil
}
```
