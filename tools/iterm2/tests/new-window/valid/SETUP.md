# Scenario

**Feature**: valid flag combinations (no conflict)

```
# user picks one mode flag (or none), no -n/-r conflict
kool iterm2 <dir> [mode-flag] -> CLI handler -> valid mode -> script built
```

## Preconditions

- All cases in this branch produce exit code 0
- All cases produce a non-empty AppleScript

## Steps

1. Run with specific flag combination
2. Assert exit code 0 and non-empty script

## Context

- None

```go
func Setup(t *testing.T, req *Request) error {
    if req.Args == nil {
        req.Args = []string{}
    }
    return nil
}
```
