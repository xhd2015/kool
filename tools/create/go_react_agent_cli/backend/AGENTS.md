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

## Section Meta Is Server-Owned

Every card's meta — `title`, `hint` (the line under the title) and `empty` (the sentence shown when the card has no rows) — lives in the server tree at `server/pagemeta/parts/<Card>.json` and is registered in `server/pagemeta/pagemeta.go`. Meta is agent-facing: the page API carries it **with** the rows (`sections[].meta` plus `empty`), so an agent that never opens a `.tsx` still learns what each card is for, empty cards included.

- Author the words in the part; never retype a title, hint or empty state in TSX or in a Go string.
- Import them in React through the `@pagemeta/<Card>.json` alias (vite + tsconfig paths), never by copying the text.
- Serve them through a page document (`GET /api/pages/home`), not by hand-rolling JSON in a handler.
- A card cannot ship without meta: page documents are built from the `pagemeta` registry, so an unregistered section is an error rather than a card without its brief.

Before finishing changes that touch a page card, verify this rule with:

```sh
go test ./server/pagemeta/ && rg -n "@pagemeta/" __PROJECT_NAME__-react/src/components
```

The first command fails when a registered card is missing `title`, `hint` or `empty`; the second proves the card imports the server-owned part.
