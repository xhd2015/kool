# Scenario

**Feature**: dry-run restore — live critical panes with unrelated ids/messages →
no already-running warn; all tabs would restore; not stamped

```
Caller
  -> seed ckpt: grok + mark (fixture ids)
  -> live fixture: other grok session + other mark message
  -> sessions restore --dry-run
  <- no "already running" warn; Would restore 2 tabs; both resumes; not stamped
```

## Steps

1. DryRun; SeedDoc grok+mark; RestoreLiveFixture=miss.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "ckpt-miss.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.RestoreLiveFixture = RestoreLiveMiss
	return nil
}
```
