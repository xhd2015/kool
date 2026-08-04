# Scenario

**Feature**: requested candidate index is focused

```text
focus --index 1 target -> candidates [0,1] -> focus candidate 1 only
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus", "--index", "1"}; return nil }
```
