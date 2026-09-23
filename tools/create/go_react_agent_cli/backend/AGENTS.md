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

## CLI: Keep the CLI Aligned with the Web

The CLI and the web page are two views of one URL space. When you add a web
page or route, add its CLI path in the same change — the CLI and the page
must always agree about what a path means.

1. One URL space: page URL ≡ CLI path ≡ server API route. `/plan/1` the page,
   `__PROJECT_NAME__ get /plan/1`, and `GET /api/plan/1` are the same address
   in three dialects.
2. Accept what the user pastes: a bare path, a web-prefixed path
   (`/<app>/plan/1`), or a full URL (strip the origin, keep the query; the
   origin pins the server). No conversion required from the caller.
3. `get` on a page path is a **human viewing the page**: the CLI performs the
   same orchestration the page performs — every API call the page makes,
   correctly orchestrated, called and rendered (breadcrumb, title, meta,
   then every section in card order with the page's own wording and empty
   states) — with only the HTML/UI markup stripped; never raw JSON. The
   replication is deliberate: super-simplified (plain text) yet still
   complete, and optimized for agents — info-dense, every line carries data.
   GET a collection path prints its rows. When the page gains a card or a
   data call, the CLI renderer gains it in the same change — an unrendered
   card is a page the agent cannot see.
4. One home for page words: the React page and the CLI renderer import the
   same text catalog, so a title or empty state cannot drift between them.
5. Writes mirror the web: a collection `put` replaces the whole list; a
   one-record `put` takes minimal fields; confirm with the resource and its
   page URL (`created plan 7 · Cubic roots` + `url …`), and map JSON-only
   sub-paths back to the page that displays them.
6. Find the running server yourself (recorded server info; `--port`/`--url`
   override; a full URL in the path wins). Never hardcode a port in docs or
   examples beyond the documented default.
7. `--json` prints the page document or raw records; `--dry-run` prints the
   request without sending; warnings go to stderr (`warning:`, exit 0),
   errors to stderr (`Error:`, non-zero); `-h/--help` at every level, and the
   root help enumerates the whole path space.
8. Images and attachments are addresses, never black holes: render a local
   path when the file is already materialized (imported/downloaded under the
   app's data dir); otherwise leave the source URL, and `get <attachment-url>`
   downloads it to a local tmp dir and prints the path, so the agent can open
   the file in its next step. Never inline binary content.

Full recipe: `go-best-practice skill --show cli/web-like-cli`. Reference
implementation: the spl repo's `ai-workshop` package (`pagepath.go`,
`pageresolve.go`, `pagetext/`, `pageview_story.go`). This template's
get/put/post/delete client ships the transport half (2, 6's flags, 7); the
page documents (3), the shared text catalog (4), write feedback (5) and
recorded-server discovery (6) bind as you add real pages.
