# Scenario

**Feature**: missing path argument shows usage error

```
# no positional path
kool vscode open -> stderr usage error
```

## Steps
1. Run `kool vscode open` with no arguments.

```go
func Setup(t *testing.T, req *Request) error {
	req.DirPath = ""
	return nil
}
```