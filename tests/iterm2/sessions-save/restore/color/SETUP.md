# Scenario

**Feature**: restore dry-run color force flags

```
Caller
  -> sessions restore --dry-run --color
  -> seeded checkpoint
  <- ANSI on Would restore / new window tokens
```

## Steps

1. ModeRestore; DryRun; seed grok+mark checkpoint (shared by color leaves).

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
	req.FilePath = "restore-color.json"
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
