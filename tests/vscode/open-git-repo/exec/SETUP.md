# Scenario

**Feature**: successful open invokes OS handler with built vscode:// URI

```
# end-to-end open path with mocked exec
OpenGitRepo -> buildGitOpenRepoURI -> OS opener(exec mock)
```

## Context
- Tests inject exec hook to capture opener command without launching VS Code.

```go
func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "exec"
	req.GoOS = "darwin"
	return nil
}

// markExecTree keeps hierarchical child packages importing this package live.
func markExecTree() {}
```