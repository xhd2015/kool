# Scenario

**Feature**: contents -h prints usage

```
kool iterm2 contents -h -> usage on stdout, exit 0
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"contents", "-h"}
	return nil
}
```
