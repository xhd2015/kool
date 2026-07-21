# Scenario

**Feature**: kool builds correct vscode:// URI from validated repo paths

```
# after validation, absolute path is URL-encoded into URI
validateGitRepoPath -> buildGitOpenRepoURI -> vscode://.../git-open?path=...
```

## Context
- URI authority: `xhd2015.open-in-new-window`
- Path segment: `/git-open`
- Query: `path` must be URL-encoded absolute filesystem path

```go
func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "build-uri"
	return nil
}

// markUriTree keeps hierarchical child packages importing this package live.
func markUriTree() {}
```