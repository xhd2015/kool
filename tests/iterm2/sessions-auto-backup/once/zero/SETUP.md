# Scenario

**Feature**: 0 critical sessions — no write / no clobber of existing auto file

```
Caller
  -> sessions auto-backup --once --file …
  -> idle-only fixture
  <- 0 critical message; file absent or previous backup kept
```
