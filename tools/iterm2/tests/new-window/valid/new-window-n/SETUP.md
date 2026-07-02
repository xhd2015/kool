# Scenario

**Feature**: new-window mode via -n short flag

```
# -n → ModeForceNew → BuildForceNewWindowScript (skip scan, always new window)
kool iterm2 <dir> -n -> ModeForceNew -> new window always
```

## Preconditions

- Args = ["-n"]

## Steps

1. Run with -n flag

## Context

- ModeForceNew is a new mode (value 2)
- Must skip session scanning entirely

```go
func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"-n"}
    return nil
}
```
