# Scenario

**Feature**: kool prechecks VS Code CLI and extension before opening

```
# precheck gate runs before IPC
ValidateDirPath -> EnsureCodeCLI -> EnsureExtensionListed -> IPC
```

## Context
- Precheck failures block IPC and OS opener.
- Tests inject fake `code` scripts via `SetCodeCommandForTest`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "cli"
	return nil
}
```