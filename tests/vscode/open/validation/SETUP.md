# Scenario

**Feature**: kool rejects invalid directory input before precheck or open

```
# validation gate prevents IPC and OS open on bad paths
kool vscode open <path> -> ValidateDirPath -> (error, no precheck)
```

## Context
- All validation failures must exit non-zero with clear stderr and never invoke precheck, IPC, or OS opener.

```go
func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "cli"
	return nil
}

// markValidationTree keeps hierarchical child packages importing this package live.
func markValidationTree() {}
```