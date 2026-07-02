# Scenario

**Feature**: new-window mode via --new-window long flag

```
# --new-window → same as -n → ModeForceNew
kool iterm2 <dir> --new-window -> ModeForceNew (via alias)
```

## Preconditions

- Args = ["--new-window"]

## Steps

1. Run with --new-window flag

## Context

- --new-window is the long form of -n

```go
func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"--new-window"}
    return nil
}
```
