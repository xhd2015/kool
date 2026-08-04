# Scenario

**Feature**: invalid candidate index offers available choices

```text
focus --index 2 target -> candidates [0,1] -> error + candidate list
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"focus", "--index", "2"}; return nil }
```
