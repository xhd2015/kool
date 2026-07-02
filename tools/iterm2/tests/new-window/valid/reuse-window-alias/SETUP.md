# Scenario

**Feature**: reuse mode via --reuse-window alias

```
# --reuse-window → same as -r → ModeReuseCurrent
kool iterm2 <dir> --reuse-window -> ModeReuseCurrent (via alias)
```

## Preconditions

- Args = ["--reuse-window"]

## Steps

1. Run with --reuse-window flag

## Context

- --reuse-window is a new alias for -r/--reuse, identical behavior

```go
func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"--reuse-window"}
    return nil
}
```
