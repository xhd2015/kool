# Scenario

**Feature**: default restore when both installs on disk prefers home; shows recorded system app when it differs

```
Caller
  -> seed ckpt: window app=/Applications/iTerm.app + grok tab
  -> RestoreAppDisk=both
  -> sessions restore --dry-run
  <- restore target ~/Applications/iTerm.app; recorded app system; Would restore; not stamped
```

## Steps

1. ModeRestore; DryRun; RestoreAppDisk=both; SeedRawJSON with system app.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.DryRun = true
	req.RestoreAppDisk = RestoreDiskBoth
	req.FilePath = "ckpt-prefer-home.json"
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
      "name": "Win-System-App",
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
