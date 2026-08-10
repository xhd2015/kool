# Scenario

**Feature**: nested doctest root for `kool iterm2 sessions save|restore` (+ Space + multi-app + restore prefer-home / `--same-app`)

```
# no live iTerm / Mission Control / dual-process; fixtures via phased + multi-app topology + disk inject
Caller
  -> kool iterm2 sessions save|restore [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space] [--spaces] [--same-app]
  -> SnapshotCollector (phased fixture) + preflight app / multi-app merge + Space inject
  -> critical filter (grok/codex/mark) + app + space record / placement plan
  -> restore app resolver (disk presence → prefer-home or --same-app per-window)
  -> plan stream / file write / restore plan (restore target / recorded app / app)
```
