# Scenario

**Feature**: -n short + --reuse long mixed conflict

```
# -n --reuse mixed flags → error
kool iterm2 -n --reuse <dir> -> error
```

## Preconditions

- Args = ["-n", "--reuse"]

## Steps

1. Run with -n short and --reuse long

## Context

- Testing mixed short/long form

```go
func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"-n", "--reuse"}
    return nil
}
```
