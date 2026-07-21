# Scenario

**Feature**: `--json` machine-readable open result on stdout

```
# JSON mode reports ipc_handled and optional fallback channel
OpenDirOptions(Json) -> stdout {"ipc_handled":...,"fallback":"uri"?}
```

## Context

- `--json` suppresses the human IPC-unreachable stderr hint when URI fallback runs.
- Default (non-`--ipc-only`) mode may still URI-fallback; JSON documents that path.

```go
func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "json"
	req.GoOS = "darwin"
	req.Json = true
	installExtensionListedPrecheck(t, req)
	return nil
}

// markJsonTree keeps hierarchical child packages importing this package live.
func markJsonTree() {}
```