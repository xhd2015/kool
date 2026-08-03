# Scenario

**Feature**: dry-run restore — same checkpoint window with hit + miss tabs;
header shows remaining (would-create) counts; both skip and would-restore lines

```
Caller
  -> seed one window: grok + mark
  -> live: only grok matches
  -> sessions restore --dry-run
  <- 1 window / 1 tab would create; skip grok line; would-restore mark line
```

## Steps

1. DryRun; SeedDoc grok+mark; RestoreLiveFixture=mixed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "ckpt-mixed.json"
	req.SeedDoc = seedGrokMarkDoc()
	req.RestoreLiveFixture = RestoreLiveMixed
	return nil
}
```
