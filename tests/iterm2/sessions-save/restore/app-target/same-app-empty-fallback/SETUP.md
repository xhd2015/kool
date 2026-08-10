# Scenario

**Feature**: `--same-app` with empty/missing app falls back to prefer-home + warning

```
Caller
  -> seed ckpt window without app field
  -> RestoreAppDisk=both; SameApp
  -> sessions restore --dry-run
  <- warning on stderr; plan falls back to home target; Would restore; not stamped
```

## Steps

1. ModeRestore; DryRun; SameApp; RestoreAppDisk=both; SeedRawJSON without app.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.DryRun = true
	req.SameApp = true
	req.RestoreAppDisk = RestoreDiskBoth
	req.FilePath = "ckpt-same-app-empty.json"
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
      "name": "Win-No-App",
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
