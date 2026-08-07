# Scenario

**Feature**: successful load-devbox merges files and emits load notices

```
./primary [--load-devbox ABS]... -- sh -c '…'
  -> exit 0; guest sees merged files; stdout may include notice: loading devbox
```

## Steps

1. Happy leaves pack primary + secondary fixtures and set guest SealedArgs.
2. Prefer `sh -c '…'` for portable guest commands.
3. Inherit `SealedDoubleDash=true` from `load-devbox/` parent unless a leaf clears it
   (e.g. stop-on-first-arg).
