# Scenario

**Feature**: ambiguous focus has no default selection

```text
focus target -> 2 candidates -> list candidates and retry guidance
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus"}; return nil }
```
