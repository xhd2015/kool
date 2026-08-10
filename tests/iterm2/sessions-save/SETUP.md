# Scenario

**Feature**: nested doctest root for `kool iterm2 sessions save|restore` (+ Space + multi-app)

```
# no live iTerm / Mission Control / dual-process; fixtures via phased + multi-app topology
Caller
  -> kool iterm2 sessions save|restore [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space] [--spaces]
  -> SnapshotCollector (phased fixture) + preflight app / multi-app merge + Space inject
  -> critical filter (grok/codex/mark) + app + space record / placement plan
  -> plan stream / file write / restore plan (restore ignores app)
```
