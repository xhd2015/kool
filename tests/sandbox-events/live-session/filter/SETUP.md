# Scenario

**Feature**: sessions discard events whose path is not in the runtime-load set

```
notify-event --path /other/not-in-session
  -> session load files unchanged
```

## Steps

1. Live session with known load path; notify a different absolute path.
