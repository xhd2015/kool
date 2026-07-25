# Scenario

**Feature**: save writes checkpoint with version, saved_at, no restored_at; grok + mark

```
Caller
  -> sessions save --file sessions-save.json
  -> critical fixture
  <- Saved; JSON version=1, restored_at empty, 2 sessions
```

## Steps

1. ModeSave; UseCriticalFixture; FilePath=sessions-save.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSave
	req.UseCriticalFixture = true
	req.FilePath = "sessions-save.json"
	return nil
}
```
