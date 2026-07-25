# Scenario

**Feature**: --json and --html are mutually exclusive

```
sessions snapshot --json --html
  -> exit non-zero; stderr mutually exclusive
  -> no fixture / capture required
```

## Steps

1. JSON and HTML both true (Run skips fixture when both set).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.JSON = true
	req.HTML = true
	// Ensure fixture not required: parent may set UseTwoWindowFixture; clear.
	req.UseTwoWindowFixture = false
	req.ObserveStreamOrder = false
	req.ITermRunning = nil
	return nil
}
```
