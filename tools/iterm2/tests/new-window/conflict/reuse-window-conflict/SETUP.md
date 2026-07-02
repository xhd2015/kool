# Scenario

**Feature**: -n --reuse-window conflict (new alias)

```
# -n --reuse-window → error (--reuse-window is alias for -r/--reuse)
kool iterm2 -n --reuse-window <dir> -> error
```

## Preconditions

- Args = ["-n", "--reuse-window"]

## Steps

1. Run with -n and --reuse-window

## Context

- --reuse-window is a new alias, should conflict with -n just like -r

```go
func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"-n", "--reuse-window"}
    return nil
}
```
