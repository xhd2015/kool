# Scenario

**Feature**: CLI accepts `--replace` flag on `kool vscode open`

```
# end-to-end subprocess with --replace before path
kool vscode open --replace <dir> -> ValidateDirPath -> IPC/URI with replace
```

## Context
- `--replace` is boolean; only valid on `open`, not `open-git-repo`.
- Flag may appear before the directory path argument.

```go
func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "cli"
	installExtensionListedPrecheck(t, req)
	return nil
}

// markCliTree keeps hierarchical child packages importing this package live.
func markCliTree() {}
```