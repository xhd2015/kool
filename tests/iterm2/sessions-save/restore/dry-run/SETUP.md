# Scenario

**Feature**: restore --dry-run prints plan; does not set restored_at

```
Caller
  -> seed ckpt.json (grok + mark)
  -> sessions restore --dry-run
  <- Would restore + resume cmds; restored_at still null
```

## Steps

1. ModeRestore; DryRun; SeedDoc with two tabs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.DryRun = true
	req.FilePath = "ckpt.json"
	req.SeedDoc = &iterm2.SaveDocument{
		Version: 1,
		SavedAt: "2026-07-25T18:00:00+0800",
		Host:    "testhost",
		Source:  "kool-iterm2-sessions-save",
		Summary: iterm2.SaveSummary{Windows: 1, Tabs: 2, Sessions: 2, ByKind: map[string]int{"grok": 1, "mark": 1}},
		Windows: []iterm2.SaveWindow{{
			SourceIndex: 1,
			Tabs: []iterm2.SaveTab{
				{Cwd: "/proj/a", Kind: "grok", SessionID: fixtureGrokSessionID, ResumeCmd: "grok --resume " + fixtureGrokSessionID},
				{Cwd: "/proj/b", Kind: "mark", Message: "still waiting for CI", ResumeCmd: "mark 'still waiting for CI'"},
			},
		}},
	}
	return nil
}
```
