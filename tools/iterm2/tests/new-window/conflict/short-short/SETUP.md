# Scenario

**Feature**: -n -r short flag conflict

```
# -n -r both set → error
kool iterm2 -n -r <dir> -> error
```

## Preconditions

- Args = ["-n", "-r"]

## Steps

1. Run with both -n and -r

## Context

- Order shouldn't matter, but testing -n before -r

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = d
    req.Args = []string{"-n", "-r"}
    return nil
}
```
