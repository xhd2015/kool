# Scenario

**Feature**: sessions restore — recreate windows from checkpoint (+ Space placement; ignores app)

```
Caller
  -> sessions restore [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space]
  -> read SaveDocument (app ignored); plan space N (Desktop N+1) unless ignore
  <- plan / apply (Switch+Create+AS) / consumed error
```
