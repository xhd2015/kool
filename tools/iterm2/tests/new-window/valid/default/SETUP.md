# Scenario

**Feature**: default mode — no mode flags

```
# no -r or -n → ModeSmart → BuildScript (scan + tab/new window)
kool iterm2 <dir> -> ModeSmart -> scan sessions -> tab or window
```

## Preconditions

- No mode flags in Args

## Steps

1. Run with empty Args

## Context

- See valid SETUP.md

```go
func Setup(t *testing.T, req *Request) error {
    if req.Args == nil {
        req.Args = []string{}
    }
    return nil
}
```
