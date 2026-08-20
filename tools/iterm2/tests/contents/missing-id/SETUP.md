# Scenario

**Feature**: missing session-id is an error

```
kool iterm2 contents -> Error: contents: missing session-id
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"contents"}
	return nil
}
```
