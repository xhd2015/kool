# Scenario

**Feature**: capture / iTerm fail is soft — warning + exit 0; no clobber

```
Caller
  -> sessions auto-backup --once --file …
  -> FailSnapshotCapture (iTerm not running)
  <- warning: on stderr; exit 0; previous file kept when seeded
```
