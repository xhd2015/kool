# Scenario

**Feature**: existing un-restored checkpoint + save (non-TTY) → error, file unchanged

```
Caller
  -> seed pending checkpoint (old-sess)
  -> sessions save (non-TTY CI default)
  <- non-zero; TTY/overwrite error; file still old-sess
```

## Steps

1. ModeSave; UseCriticalFixture; SeedDoc pending; FilePath=pending.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSave
	req.UseCriticalFixture = true
	req.FilePath = "pending.json"
	req.SeedDoc = &iterm2.SaveDocument{
		Version: 1,
		SavedAt: "2026-07-01T00:00:00+0800",
		Host:    "old",
		Source:  "kool-iterm2-sessions-save",
		Summary: iterm2.SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []iterm2.SaveWindow{{
			SourceIndex: 1,
			Tabs: []iterm2.SaveTab{{
				Cwd: "/old", Kind: "grok", SessionID: "old-sess", ResumeCmd: "grok --resume old-sess",
			}},
		}},
	}
	return nil
}
```
