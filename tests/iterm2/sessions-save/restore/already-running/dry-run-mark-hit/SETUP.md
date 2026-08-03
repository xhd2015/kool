# Scenario

**Feature**: dry-run restore — live mark exact `message` match → warn + skip;
remaining grok would restore; not stamped

```
Caller
  -> seed ckpt: grok + mark (fixtureMarkMessage)
  -> live fixture: only mark with same message
  -> sessions restore --dry-run
  <- warning already running + mark message + pid; skip mark; grok would restore
```

## Steps

1. DryRun; SeedDoc grok+mark; RestoreLiveFixture=mark-hit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "ckpt-mark-hit.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.RestoreLiveFixture = RestoreLiveMarkHit
	return nil
}
```
