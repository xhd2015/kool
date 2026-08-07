# Scenario

**Feature**: build with --runtime-load-devbox seals absolute remote paths

```
# runtime-load-devbox pack branch
user -> kool sandbox build -o OUT --runtime-load-devbox ABS... [--file|--env]...
  -> seal path strings into PackBlob.runtime_load_devbox (no local open)
  -> inspect lists sealed absolute paths
```

## Steps

1. Leaves set `RuntimeLoadDevbox` entries and fixtures.
2. Success leaves still pass ≥1 `--file` or `--env` (load paths alone do not fill pack).
