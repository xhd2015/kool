# Scenario

**Feature**: kool opens directories via IPC before URI fallback

```
# IPC client sends open op with normalized absolute path
OpenDir -> IPC client {"op":"open","path":"/abs/dir"} -> VS Code extension
```

## Context
- Mock Unix socket server accepts JSON-line requests and returns `{"ok":true}`.
- OS opener must not run when IPC succeeds.
- Tests override socket path via `SetIPC_SOCKETPathForTest`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "ipc"
	req.GoOS = "darwin"
	installExtensionListedPrecheck(t, req)
	return nil
}
```