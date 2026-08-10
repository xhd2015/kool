# Scenario

**Feature**: restore ignores checkpoint `app`; dry-run still plans cd + resume

```
Caller
  -> seed ckpt-with-app.json including "app" on window
  -> sessions restore --dry-run
  <- Would restore; no app placement / no preferred-app targeting; not stamped
```

## Steps

1. ModeRestore; DryRun; SeedRawJSON with app + one grok tab.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.DryRun = true
	req.FilePath = "ckpt-with-app.json"
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
      "name": "Win-With-App",
      "app": "~/Applications/iTerm.app",
      "space": 1,
      "iterm_window_id": 508113,
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
