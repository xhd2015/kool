# Scenario

**Feature**: live restore — one tab already running → warn; AppleScript only for
remaining tabs; stamp restored_at; summary includes skip count

```
Caller
  -> seed ckpt: grok + mark
  -> live: only grok matches
  -> MockRestoreAS + Space backend
  -> sessions restore (live)
  <- warn; AS has mark resume not grok; restored_at set; skip in summary
```

## Steps

1. Live (not dry-run); SeedDoc; RestoreLiveFixture=partial-skip; MockRestoreAS.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = false
	req.FilePath = "ckpt-live-partial.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.RestoreLiveFixture = RestoreLivePartial
	req.MockRestoreAS = true
	return nil
}
```
