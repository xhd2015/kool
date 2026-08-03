# Scenario

**Feature**: restore --dry-run shows recorded space 2 as Desktop 3; not stamped

```
Caller
  -> seed ckpt.json with space: 2 (and optional iterm_window_id info-only)
  -> sessions restore --dry-run
  <- plan includes space 2 (Desktop 3); no iterm_window_id text; restored_at null
```

## Steps

1. FilePath=ckpt-space2.json; SeedRawJSON with space 2 + one grok tab.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FilePath = "ckpt-space2.json"
	req.SeedRawJSON = `{
  "version": 1,
  "saved_at": "2026-07-25T18:00:00+0800",
  "restored_at": null,
  "host": "testhost",
  "source": "kool-iterm2-sessions-save",
  "summary": {
    "windows": 1,
    "tabs": 1,
    "sessions": 1,
    "by_kind": { "grok": 1 }
  },
  "windows": [
    {
      "source_index": 1,
      "space": 2,
      "iterm_window_id": 4242,
      "tabs": [
        {
          "cwd": "/proj/a",
          "kind": "grok",
          "session_id": "` + fixtureGrokSessionID + `",
          "resume_cmd": "grok --resume ` + fixtureGrokSessionID + `"
        }
      ]
    }
  ]
}
`
	return nil
}
```
