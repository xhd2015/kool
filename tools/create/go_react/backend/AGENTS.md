# Agent Guidelines

## React: Never Hide Content During Loading

Never use the pattern `loading ? <Loading/> : <MainContent/>` because it hides already-loaded content during refresh.

Instead, show loading indicators alongside existing content:

```tsx
// Bad: hides content during refresh
{loading ? <div>Loading...</div> : <div>{content}</div>}

// Good: show loading only when no content exists yet
{loading && content == null && <div>Loading...</div>}
{content != null && <div>{content}</div>}
```

When refreshing data that is already displayed:
- Do NOT clear existing state before fetching
- Show a subtle loading indicator without removing visible content
- Replace content silently once new data arrives

## Frontend: API Calls Stay Under `__PROJECT_NAME__-react/src/api`

Do not call `fetch`, `apiFetch`, `new EventSource`, or `apiEventSource` outside `__PROJECT_NAME__-react/src/api`.

All frontend API transport and endpoint wrappers must live under `__PROJECT_NAME__-react/src/api`. Components, hooks, and feature pages should import typed API functions from that directory instead of constructing API requests directly.

Before finishing frontend API changes, verify this rule with:

```sh
rg -n "\bfetch\(|\bapiFetch\b|new EventSource|\bapiEventSource\b" __PROJECT_NAME__-react/src --glob '!__PROJECT_NAME__-react/src/api/**'
```

The command should return no matches.
