# Scenario

**Feature**: `--same-app` recreates window in recorded system app (dry-run plan + live path-tell AS)

```
Caller
  -> seed ckpt app=/Applications/iTerm.app
  -> RestoreAppDisk=both; SameApp; MockRestoreAS live
  -> sessions restore (live)
  <- AS path-tells system install; stamp restored_at
```

## Steps

1. ModeRestore; SameApp; RestoreAppDisk=both; MockRestoreAS; SeedRawJSON system app.
2. Live (not dry-run) so MockRestoreAS captures path-tell scripts.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.DryRun = false
	req.SameApp = true
	req.RestoreAppDisk = RestoreDiskBoth
	req.MockRestoreAS = true
	req.FilePath = "ckpt-same-app-system.json"
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
      "name": "Win-Same-System",
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
