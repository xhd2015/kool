# Scenario

**Feature**: dry-run restore — codex `kind`+`session_id` match → warn + skip;
not confused with grok (different kind key)

```
Caller
  -> seed ckpt: one codex tab (fixtureCodexSessionID)
  -> live: codex same session_id
  -> sessions restore --dry-run
  <- already-running warn with codex id; 0 would create; not stamped
```

## Steps

1. DryRun; SeedDoc single codex tab; RestoreLiveFixture=codex-hit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.FilePath = "ckpt-codex.json"
	req.SeedDoc = &iterm2.SaveDocument{
		Version: 1,
		SavedAt: "2026-07-25T18:00:00+0800",
		Host:    "testhost",
		Source:  "kool-iterm2-sessions-save",
		Summary: iterm2.SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"codex": 1}},
		Windows: []iterm2.SaveWindow{{
			SourceIndex: 1,
			Name:        "Win-Codex",
			Tabs: []iterm2.SaveTab{{
				Name:      "codex-tab",
				Cwd:       "/proj/codex",
				Kind:      "codex",
				SessionID: fixtureCodexSessionID,
				ResumeCmd: "codex resume " + fixtureCodexSessionID,
			}},
		}},
	}
	req.RestoreLiveFixture = RestoreLiveCodexHit
	return nil
}
```
