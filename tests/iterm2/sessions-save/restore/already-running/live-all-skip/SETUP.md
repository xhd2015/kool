# Scenario

**Feature**: live restore — all tabs already running → still stamp restored_at;
exit 0; no create-window AppleScript (AS not called / no create)

```
Caller
  -> seed ckpt: grok + mark
  -> live: both match
  -> MockRestoreAS
  -> sessions restore (live)
  <- warn both; restored_at set; AS call count 0 or no create window; 0 restored + N skipped
```

## Steps

1. Live; SeedDoc; RestoreLiveFixture=all-skip; MockRestoreAS.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = false
	req.FilePath = "ckpt-live-all-skip.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.RestoreLiveFixture = RestoreLiveAllSkip
	req.MockRestoreAS = true
	return nil
}
```
