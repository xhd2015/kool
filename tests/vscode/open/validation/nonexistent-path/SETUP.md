# Scenario

**Feature**: nonexistent path fails before precheck

```
# path does not exist on disk
ValidateDirPath(nonexistent) -> error
```

## Steps
1. Run CLI with a path that does not exist.

```go
func Setup(t *testing.T, req *Request) error {
	markValidationTree()
	markRootTree()
	req.DirPath = "/tmp/kool-vscode-open-does-not-exist-xyz"
	return nil
}
```