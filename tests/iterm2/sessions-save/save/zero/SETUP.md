# Scenario

**Feature**: idle-only snapshot → 0 critical, no file

```
Caller
  -> sessions save --file empty.json
  -> idle-only fixture
  <- "0 critical"; empty.json not created
```

## Steps

1. ModeSave; UseIdleOnlyFixture; FilePath=empty.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSave
	req.UseIdleOnlyFixture = true
	req.FilePath = "empty.json"
	return nil
}
```
