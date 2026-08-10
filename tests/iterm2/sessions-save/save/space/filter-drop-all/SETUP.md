# Scenario

**Feature**: save `--spaces 1` drops all critical windows (on 0 and 2); no write

```
Caller
  -> sessions save --spaces 1 --file filter-drop-all.json
  -> two critical windows (space 0 + space 2)
  <- 0 critical; skip warn; no file write
```

## Steps

1. UseTwoCriticalSpacesFixture; Spaces=1; FilePath=filter-drop-all.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseCriticalFixture = false
	req.UseTwoCriticalSpacesFixture = true
	req.Spaces = "1"
	req.FilePath = "filter-drop-all.json"
	return nil
}
```
