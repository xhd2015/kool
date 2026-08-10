# Scenario

**Feature**: default restore when only home install on disk targets home

```
Caller
  -> seed ckpt with system recorded app
  -> RestoreAppDisk=home
  -> sessions restore --dry-run
  <- restore target ~/Applications/iTerm.app; not system as target; not stamped
```

## Steps

1. ModeRestore; DryRun; RestoreAppDisk=home; SeedRawJSON with system app.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.DryRun = true
	req.RestoreAppDisk = RestoreDiskHome
	req.FilePath = "ckpt-only-home.json"
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
      "name": "Win-Only-Home",
      "app": "/Applications/iTerm.app",
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
