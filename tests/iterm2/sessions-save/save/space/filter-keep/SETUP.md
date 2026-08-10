# Scenario

**Feature**: save `--spaces 0` keeps only windows on space 0; records filter; skip warning

```
Caller
  -> sessions save --spaces 0 --file filter-keep.json
  -> two critical windows (space 0 + space 2)
  <- FileJSON has filter.spaces [0], one window; stderr skip warn
```

## Steps

1. UseTwoCriticalSpacesFixture; Spaces=0; FilePath=filter-keep.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseCriticalFixture = false
	req.UseTwoCriticalSpacesFixture = true
	req.Spaces = "0"
	req.FilePath = "filter-keep.json"
	return nil
}
```
