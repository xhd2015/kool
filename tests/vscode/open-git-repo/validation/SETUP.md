# Scenario

**Feature**: kool rejects invalid input before opening vscode:// URI

```
# validation gate prevents OS open on bad paths
kool vscode open-git-repo <path> -> validateGitRepoPath -> (error, no open)
```

## Context
- All validation failures must exit non-zero with clear stderr and never invoke OS opener.

```go
func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "cli"
	return nil
}

// markValidationTree keeps hierarchical child packages importing this package live.
func markValidationTree() {}
```