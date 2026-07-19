# Scenario

**Feature**: existing file + --save --force overwrites and prints diff buckets

```
existing bots.json + run bots --tab … --save --force
  -> overwrite file; stdout/stderr mentions modified/added/deleted/unchanged
  -> no iTerm
```

## Steps

1. Write bots fixture (ids a,b; window local-bots).
2. Ad-hoc tabs: keep a unchanged command, change b command, add c, drop nothing
   on a — and change window_name for a modified window bucket.
   Actually: a unchanged, b modified command, c added → deleted none.
   To exercise deleted: omit b.
   Tabs: `[id=a] echo a` (unchanged), `[id=c] echo c` (added), b deleted;
   WindowName new-win (window_name modified).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.Save = true
	req.Force = true
	req.WindowName = "new-win"
	req.Tabs = []string{
		"[id=a] echo a",          // unchanged command
		"[id=b] echo B-changed",  // modified command
		"[id=c] echo c",          // added
	}
	return nil
}
```
