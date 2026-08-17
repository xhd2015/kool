# Scenario

**Feature**: full orchestration falls back to OS handler when IPC unavailable

```
# end-to-end OpenDir with IPC disabled
OpenDir -> precheck ok -> IPC fail -> OS opener(exec mock)
```

## Context
- Tests inject exec hook to capture opener command without launching VS Code.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "orchestrate"
	req.GoOS = "darwin"
	req.IPCAlwaysFail = true
	installExtensionListedPrecheck(t, req)
	return nil
}
```