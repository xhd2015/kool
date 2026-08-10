# Scenario

**Feature**: dual-running same window ids → no doubles + warning; exit 0 (partial)

```
Caller
  -> sessions save --file app-dedupe.json
  -> multi-app dedupe fixture (shared iterm_window_id across sources)
  <- one window; stderr collapse/dual warn; optional app set; exit 0
```

## Steps

1. UseMultiAppDedupeFixture; FilePath=app-dedupe.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseMultiAppDedupeFixture = true
	req.FilePath = "app-dedupe.json"
	return nil
}
```
