# Scenario

**Feature**: successful open invokes OS handler with built vscode:// URI

```
# end-to-end open path with mocked exec
OpenGitRepo -> buildGitOpenRepoURI -> OS opener(exec mock)
```

## Context
- Tests inject exec hook to capture opener command without launching VS Code.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "exec"
	req.GoOS = "darwin"
	installExtensionListedPrecheck(t, req)
	return nil
}
```