# Scenario

**Feature**: pre-loop validation errors for auto-backup

```
Caller
  -> sessions auto-backup --once --interval <invalid>
  <- Error: + non-zero exit before loop
```
