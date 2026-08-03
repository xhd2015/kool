# Scenario

**Feature**: dry-run restore — live grok `kind`+`session_id` match → warn + skip;
remaining mark tab would restore; not stamped

```
Caller
  -> seed ckpt: grok (fixtureGrokSessionID) + mark
  -> live fixture: only grok same session_id
  -> sessions restore --dry-run
  <- warning already running + pid for grok; skip line; mark would restore;
     header remaining counts; skipped meta; restored_at null
```

## Steps

1. DryRun; SeedDoc grok+mark; RestoreLiveFixture=agent-hit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "ckpt-agent-hit.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.RestoreLiveFixture = RestoreLiveAgentHit
	return nil
}
```
