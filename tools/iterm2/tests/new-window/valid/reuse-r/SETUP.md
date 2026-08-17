# Scenario

**Feature**: reuse mode via -r short flag

```
# -r → ModeReuseCurrent → BuildReuseCurrentSessionScript (scan + focus)
kool iterm2 <dir> -r -> ModeReuseCurrent -> scan sessions -> focus existing
```

## Preconditions

- Args = ["-r"]

## Steps

1. Run with -r flag

## Context

- None

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = d
    req.Args = []string{"-r"}
    return nil
}
```
