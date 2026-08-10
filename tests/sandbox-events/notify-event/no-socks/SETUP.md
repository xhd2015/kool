# Scenario

**Feature**: notify-event with no subscribers

```
kool sandbox notify-event --type devbox.updated --path ABS --root ROOT
  # ROOT/events empty
  -> warning on stderr; exit 0
```

## Steps

1. Ensure empty events dir under EventRoot; set path to a dummy absolute path.
