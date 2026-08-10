# Scenario

**Feature**: single-cycle auto-backup (`--once`) write / zero / soft-fail / dry-run

```
Caller
  -> sessions auto-backup --once [--file] [--dry-run] …
  -> fixture collector (critical | idle | fail-capture)
  -> one cycle: write | no-clobber | warning | plan-only
```
