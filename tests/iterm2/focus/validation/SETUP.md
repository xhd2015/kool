# Scenario

**Feature**: focus validation fails before live discovery

```text
Invalid focus arguments -> CLI validation error -> no FocusFake discovery
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Target = false; return nil }
```
