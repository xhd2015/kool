# Scenario

**Feature**: load-devbox env merge and hard conflicts across sandboxes

```
# disjoint keys merge; same key from two sandboxes → Error incompatible
./primary --load-devbox ABS -- sh -c 'printf …'
  -> exit 0 with both values | non-zero Error: on conflict
```

## Steps

1. Env leaves pack primary env and secondary env via SecondaryPacks.
2. Conflict leaves expect non-zero sealed exit and Error: on RunStderr.
