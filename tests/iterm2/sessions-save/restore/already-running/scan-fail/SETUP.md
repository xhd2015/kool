# Scenario

**Feature**: dry-run restore — live snapshot capture fails soft → warn; treat as
0 hits; full would-restore plan; not stamped (D6)

```
Caller
  -> seed ckpt: grok + mark
  -> FailSnapshotCapture (iTerm not running fixture)
  -> sessions restore --dry-run
  <- soft warning about scan/capture; no already-running skip; full plan; not stamped
```

## Steps

1. DryRun; SeedDoc; FailSnapshotCapture=true (no RestoreLiveFixture).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "ckpt-scan-fail.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.FailSnapshotCapture = true
	return nil
}
```
