# Scenario

**Feature**: live session exposes events/<session-id>.sock under KOOL_SANDBOX_ROOT

```
# with runtime-load sealed
./sandbox.bin -- <long guest>
  -> events/<id>.sock exists while guest runs
```

## Steps

1. Primary with --runtime-load-devbox secondary; long guest; StopGuest after poll.
