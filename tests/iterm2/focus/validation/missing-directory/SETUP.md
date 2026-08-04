# Scenario

**Feature**: missing target directory

```text
kool iterm2 focus -> error
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus"}; return nil }
```
