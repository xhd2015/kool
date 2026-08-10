# Scenario

**Feature**: sessions restore — recreate windows from checkpoint (+ Space placement; prefer-home / `--same-app` path target)

```
Caller
  -> sessions restore [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space] [--same-app]
  -> read SaveDocument; resolve restore target (disk inject) or per-window app
  -> plan space N (Desktop N+1) unless ignore; restore target / recorded app / app meta
  <- plan / apply (path-tell AS) / consumed error
```
