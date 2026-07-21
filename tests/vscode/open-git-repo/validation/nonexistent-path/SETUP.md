# Scenario

**Feature**: nonexistent path fails before open

```
# path does not exist on disk
validateGitRepoPath(nonexistent) -> error
```

## Steps
1. Run CLI with a path that does not exist.

```go
func Setup(t *testing.T, req *Request) error {
	markValidationTree()
	markRootTree()
	req.Phase = "cli"
	req.RepoPath = "/tmp/kool-open-git-repo-does-not-exist-xyz"
	return nil
}
```