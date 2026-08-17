# Scenario

**Feature**: missing path argument shows usage error

```
# no positional path
kool vscode open-git-repo -> stderr usage error
```

## Steps
1. Run `kool vscode open-git-repo` with no arguments.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "cli"
	req.RepoPath = ""
	return nil
}
```