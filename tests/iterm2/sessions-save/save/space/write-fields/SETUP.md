# Scenario

**Feature**: save writes checkpoint that always emits `space` (including 0) when not ignore

```
Caller
  -> sessions save --file space-save.json
  -> critical fixture; Space recording on
  <- FileJSON contains "space" key; exit 0; Saved
```

## Steps

1. FilePath=space-save.json (live write, not dry-run).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FilePath = "space-save.json"
	return nil
}
```
