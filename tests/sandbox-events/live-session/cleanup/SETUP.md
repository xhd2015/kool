# Scenario

**Feature**: session end closes listener and unlinks the event sock

```
# live guest exits
  -> events/<id>.sock unlinked
```

## Steps

1. Short sleep guest or ready/stop; WaitGuestExit; assert sock gone.
