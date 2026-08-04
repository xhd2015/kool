# Scenario

**Feature**: focus-specific help documents selection

```text
kool iterm2 focus --help -> focus usage
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus", "--help"}; return nil }
```
