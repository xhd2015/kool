# Scenario

**Feature**: missing path argument shows usage error

```
# no positional path
kool vscode open -> stderr usage error
```

## Steps
1. Run `kool vscode open` with no arguments.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DirPath = ""
	return nil
}
```