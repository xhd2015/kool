# Scenario

**Feature**: root -h mentions contents

```
kool iterm2 -h -> contents <session-id>
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```
