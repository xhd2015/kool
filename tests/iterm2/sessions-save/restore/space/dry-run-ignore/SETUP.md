# Scenario

**Feature**: restore --dry-run --ignore-macos-space omits space lines; no placement

```
Caller
  -> seed ckpt with space: 2
  -> sessions restore --dry-run --ignore-macos-space
  <- Would restore + resume; no space N (Desktop …) lines; not stamped
```

## Steps

1. IgnoreMacOSSpace; FilePath=ckpt-ignore.json; SeedRawJSON with space 2.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.IgnoreMacOSSpace = true
	req.FilePath = "ckpt-ignore.json"
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
      "iterm_window_id": 99,
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
