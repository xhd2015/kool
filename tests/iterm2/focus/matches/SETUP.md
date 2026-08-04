# Scenario

**Feature**: selecting exact directory matches

```text
kool iterm2 focus <dir> [--index N] -> FocusFake discovers -> candidate selection -> FocusFake focuses
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus"}; req.Target = true; return nil }
```
