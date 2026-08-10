# Scenario

**Feature**: dual-app merge writes both canonical apps; no duplicate window ids

```
Caller
  -> sessions save --file app-dual.json
  -> multi-app fixture: system W1 + home W2 (distinct iterm_window_ids)
  <- FileJSON has both apps; summary 2 windows / 2 sessions; no dup ids
```

## Steps

1. UseMultiAppFixture; FilePath=app-dual.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseMultiAppFixture = true
	req.FilePath = "app-dual.json"
	return nil
}
```
