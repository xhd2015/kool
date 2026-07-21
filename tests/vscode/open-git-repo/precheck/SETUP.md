# Scenario

**Feature**: kool prechecks VS Code CLI and extension before opening git repo

```
# precheck gate runs after validation, before IPC
ValidateGitRepoPath -> EnsureCodeCLI -> EnsureExtensionListed -> IPC
```

## Context
- Precheck failures block IPC and OS opener.

```go
func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "cli"
	return nil
}

// markPrecheckTree keeps hierarchical child packages importing this package live.
func markPrecheckTree() {}
```