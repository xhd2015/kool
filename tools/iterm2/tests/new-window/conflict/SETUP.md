# Scenario

**Feature**: mutual exclusion validation — -n and -r together produce an error

```
# -n and -r conflict → error: cannot specify both
kool iterm2 -n -r <dir> -> CLI handler -> error, no script
```

## Preconditions

- All cases in this branch produce exit code 1
- None produce an AppleScript (error before library call)

## Steps

1. Run with conflicting flags
2. Assert exit code 1 and non-empty stderr

## Context

- The error message should mention the conflict (mutually exclusive)

```go
func Setup(t *testing.T, req *Request) error {
    if req.Args == nil {
        req.Args = []string{}
    }
    return nil
}
```
