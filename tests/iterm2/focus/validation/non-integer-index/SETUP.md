# Scenario

**Feature**: malformed index rejected by parser

```text
kool iterm2 focus --index no <dir> -> parse error
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus", "--index", "no"}; req.Target = true; return nil }
```
