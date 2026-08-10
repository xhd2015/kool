# Scenario

**Feature**: nested doctest root for `kool iterm2 sessions auto-backup` (periodic critical-tab checkpoint)

```
# no live iTerm / launchd; fixtures via phased collector + disk seed; --once only
Caller
  -> kool iterm2 sessions auto-backup [--once] [--interval] [--file] [--dry-run] …
  -> SnapshotCollector (phased fixture) + critical filter (grok/codex/mark)
  -> always-overwrite auto file on non-empty save / soft-fail / zero no-clobber
  -> plan stream / file write / warning
```
