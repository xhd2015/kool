# Scenario

**Feature**: `kool go modules --list` streams `<dir> <path>` lines, applying scan skip rules

```
# each list leaf builds a workspace; --list delegates to scan.ScanStream and prints lines
workspace (go.mod files + git) -> kool go modules --list --dir <root> -> stdout lines
```

The `list/` siblings are MECE over the workspace layout: a clean multi-module workspace
(basic), a workspace with a nested separate repo (must be absent), and a workspace with
`testdata` (must be absent). All leaves share a root module `some.com/root` under one git
repo; each leaf adds the layout specific to its scenario.

## Steps

1. Create an isolated workspace with root `go.mod` (`some.com/root`), git-init'd.
2. Set `req.RootDir` to the workspace.
3. Leaf `Setup` adds the sub-directories specific to the scenario.

```go
// initListRoot creates an isolated workspace with a root go.mod (some.com/root) and inits
// it as a git repo, returning the workspace dir. Shared by all list leaves.
func initListRoot(t *testing.T) string {
	t.Helper()
	ws := newWorkspace(t)
	writeModule(t, ws, "some.com/root")
	initGitRepo(t, ws)
	return ws
}

func Setup(t *testing.T, req *Request) error {
	// shared root workspace; leaves add scenario-specific sub-dirs to req.RootDir
	req.RootDir = initListRoot(t)
	return nil
}
```
