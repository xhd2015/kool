# Scenario

**Feature**: --new-window --reuse long flag conflict

```
# --new-window --reuse both set → error
kool iterm2 --new-window --reuse <dir> -> error
```

## Preconditions

- Args = ["--new-window", "--reuse"]

## Steps

1. Run with both long flags

## Context

- Testing long form of both flags

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = d
    req.Args = []string{"--new-window", "--reuse"}
    return nil
}
```
