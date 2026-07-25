# Scenario

**Feature**: ListWindows returns fixture window headers

```
InstallPhasedFixtureCollectorForTest(two windows)
  -> c.ListWindows()
  -> [{Index:1, Name:Win-A}, {Index:2, Name:Win-B}]
```

## Steps

1. Mode=list-windows.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeListWindows
	return nil
}
```
