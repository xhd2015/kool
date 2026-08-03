# Scenario

**Feature**: restore --dry-run clamps space >= 16 to 0; warns; plans Desktop 1

```
Caller
  -> seed ckpt with space: 16
  -> sessions restore --dry-run
  <- plan space 0 (Desktop 1); stderr warning; not stamped
```

## Steps

1. FilePath=ckpt-space16.json; SeedRawJSON with space 16.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FilePath = "ckpt-space16.json"
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
      "space": 16,
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
