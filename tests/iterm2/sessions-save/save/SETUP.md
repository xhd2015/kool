# Scenario

**Feature**: sessions save — checkpoint critical tabs (+ optional Space fields)

```
Caller
  -> sessions save [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space]
  -> phased capture + critical filter + SpaceIndexForWindow (unless ignore)
  -> stream window plan / write JSON / error
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}
```
