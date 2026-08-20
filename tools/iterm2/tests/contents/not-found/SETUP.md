# Scenario

**Feature**: unknown session-id is Error: session not found

```
kool iterm2 contents 00000000-0000-0000-0000-000000000000
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"contents", "00000000-0000-0000-0000-000000000000"}
	req.NotFound = true
	return nil
}
```
