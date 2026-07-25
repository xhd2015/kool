# Scenario

**Feature**: restore when restored_at already set → error

```
Caller
  -> seed consumed.json (restored_at set)
  -> sessions restore
  <- non-zero; consumed / restored_at error
```

## Steps

1. ModeRestore; SeedDoc with RestoredAt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.FilePath = "consumed.json"
	ts := "2026-07-25T20:00:00+0800"
	req.SeedDoc = &iterm2.SaveDocument{
		Version:    1,
		SavedAt:    "2026-07-25T18:00:00+0800",
		RestoredAt: &ts,
		Host:       "testhost",
		Source:     "kool-iterm2-sessions-save",
		Summary:    iterm2.SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []iterm2.SaveWindow{{
			SourceIndex: 1,
			Tabs: []iterm2.SaveTab{
				{Cwd: "/a", Kind: "grok", SessionID: "x", ResumeCmd: "grok --resume x"},
			},
		}},
	}
	return nil
}
```
