# Scenario

**Feature**: notify-event --dry-run lists targets without requiring live listeners

```
kool sandbox notify-event --type devbox.updated --path ABS --root ROOT --dry-run
  -> lists would-be sock targets; exit 0
```

## Steps

1. Create placeholder `*.sock` file names under events/ (not necessarily listening).
