# Scenario

**Feature**: kool builds correct vscode:// URI from validated directory paths

```
# after validation, absolute path is URL-encoded into URI
ValidateDirPath -> BuildOpenURI -> vscode://.../open?path=...
```

## Context
- URI authority: `xhd2015.open-in-new-window`
- Path segment: `/open`
- Query: `path` must be URL-encoded absolute filesystem path; `replace=true` only when flagged

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "build-uri"
	return nil
}
```