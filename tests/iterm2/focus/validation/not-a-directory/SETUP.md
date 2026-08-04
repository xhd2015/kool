# Scenario

**Feature**: file target rejected

```text
kool iterm2 focus <file> -> validation error
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus"}; req.Target = true; req.TargetKind = "file"; return nil }
```
