# Scenario

**Feature**: focus help is available at both iTerm2 command levels

```text
User -> kool iterm2 help or focus help -> usage text
```

## Steps

1. Invoke the requested help form.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Target = false; return nil }
```
