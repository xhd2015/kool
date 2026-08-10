# Scenario

**Feature**: `--same-app` with two windows different recorded apps shows each create target

```
Caller
  -> seed W1 system app + W2 home app
  -> RestoreAppDisk=both; SameApp
  -> sessions restore --dry-run
  <- W1 app system; W2 app home; Would restore; not stamped
```

## Steps

1. ModeRestore; DryRun; SameApp; RestoreAppDisk=both; SeedRawJSON mixed apps.

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
	req.FilePath = "ckpt-same-app-mixed.json"
	req.SeedRawJSON = `{
  "version": 1,
  "saved_at": "2026-07-25T18:00:00+0800",
  "restored_at": null,
  "host": "testhost",
  "source": "kool-iterm2-sessions-save",
  "summary": {
    "windows": 2,
    "tabs": 2,
    "sessions": 2,
    "by_kind": { "grok": 1, "mark": 1 }
  },
  "windows": [
    {
      "source_index": 1,
      "name": "Win-System",
      "app": "/Applications/iTerm.app",
      "tabs": [
        {
          "cwd": "/proj/system",
          "kind": "grok",
          "session_id": "` + fixtureGrokSessionID + `",
          "resume_cmd": "grok --resume ` + fixtureGrokSessionID + `"
        }
      ]
    },
    {
      "source_index": 2,
      "name": "Win-Home",
      "app": "~/Applications/iTerm.app",
      "tabs": [
        {
          "cwd": "/proj/home",
          "kind": "mark",
          "message": "` + fixtureMarkMessage + `",
          "resume_cmd": "mark '` + fixtureMarkMessage + `'"
        }
      ]
    }
  ]
}
`
	return nil
}
```
