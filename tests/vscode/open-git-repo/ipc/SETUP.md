# Scenario

**Feature**: kool opens git repos via IPC before URI fallback

```
# IPC client sends git-open op with normalized absolute repo path
OpenGitRepo -> IPC client {"op":"git-open","path":"/abs/repo"} -> VS Code extension
```

## Context
- Mock Unix socket server returns `{"ok":true}`.
- OS opener must not run when IPC succeeds.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "ipc"
	req.GoOS = "darwin"
	installExtensionListedPrecheck(t, req)
	return nil
}
```