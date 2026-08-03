# Scenario

**Feature**: dry-run restore — all checkpoint tabs already live → 0 would create;
all skip lines; not stamped (dry-run never stamps)

```
Caller
  -> seed ckpt: grok + mark
  -> live fixture: both match
  -> sessions restore --dry-run
  <- 0 windows/tabs would restore; skip lines for both; skip meta; not stamped
```

## Steps

1. DryRun; SeedDoc grok+mark; RestoreLiveFixture=all-skip.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "ckpt-all-skip.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.RestoreLiveFixture = RestoreLiveAllSkip
	return nil
}
```
