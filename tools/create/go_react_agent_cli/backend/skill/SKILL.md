---
name: __PROJECT_NAME__
description: Agent-driven web app - a Go API server plus a React UI. Drive the running __PROJECT_NAME__ server from the CLI - start it, then get/put/post/delete its API to inspect and change data. Use when the user asks __PROJECT_NAME__ to show, plan, or illustrate something, or mentions the __PROJECT_NAME__ server, its API, or its CLI.
metadata:
  version: 0.1.0
---

# __PROJECT_NAME__

__PROJECT_NAME__ is an agent-driven web app: a Go API server plus a React UI.
Agents interact with it through its CLI: one command starts the server, the
HTTP verb commands read and change its data, and every API response is JSON.

## Start the server

```sh
__PROJECT_NAME__ server            # listens on the first free port from 8080
__PROJECT_NAME__ server --dev      # dev mode: proxies to the Vite dev server (HMR)
```

The ready line prints the chosen port, e.g. `__PROJECT_NAME__ ready: http://localhost:8080/`.

## Talk to the server

URI forms: a bare path talks to `http://localhost:8080`; a full URL
(`http://localhost:9000/x`) is used as-is. `--port <n>` / `--url <origin>`
override the target.

```sh
__PROJECT_NAME__ get <URI>                                  # print the response body
__PROJECT_NAME__ put <URI> <json|@file|-> [--json]          # replace a resource
__PROJECT_NAME__ post <URI> [<json|@file|->] [--json]       # create a resource
__PROJECT_NAME__ delete <URI> [--json]                      # delete a resource
```

- Writes are silent on success; pass `--json` to print the response body.
- Non-2xx responses are errors on stderr with the status and body; exit code 1.
- `--dry-run` (put/post/delete) prints the request without sending it.
- A body is a JSON literal, `@file`, or `-` for stdin.

## Endpoints

| Method | Path             | Purpose                                        |
|--------|------------------|------------------------------------------------|
| GET    | /ping            | health check, replies `pong`                   |
| GET    | /api/counter     | current counter (`{"last":N}`)                 |
| POST   | /api/counter     | allocate the next id (`{"id":N+1}`)            |
| GET    | /api/page-meta   | every card's server-owned meta (title/hint/empty) |
| GET    | /api/pages/home  | the Home page document: cards in web order, each with `meta`, `empty` and its rows |

Card meta is server-owned: a page document carries `sections[].meta`
(`title`, `hint`, `empty`) next to the rows, so `get /api/pages/home` tells you
what every card is for — empty cards included — without reading the React
source. Author the words in `server/pagemeta/parts/<Card>.json`; the React card
imports that same file over the `@pagemeta/` alias, and
`go test ./server/pagemeta/` fails when a registered card is missing a word.

The counter demonstrates persistent state: ids come from
`~/.<project>/id.json` and survive restarts. Grow the API on top of this
pattern: one JSON handler per route, registered in `server.RegisterAPI`.

## Walkthrough

```sh
__PROJECT_NAME__ server &                     # start (or: go run ./cmd/__PROJECT_NAME__ server)
__PROJECT_NAME__ get /ping                    # pong
__PROJECT_NAME__ post /api/counter            # allocate id 1 (silent)
__PROJECT_NAME__ get /api/counter --json      # {"last":1}
__PROJECT_NAME__ get /api/pages/home --json   # cards with meta + rows + empty flags
__PROJECT_NAME__ get http://localhost:8080/api/counter   # full-URL form
```

## Tips

- Check the server is up with `__PROJECT_NAME__ get /ping` before other calls.
- Use `--dry-run` to show an agent's intended write before committing to it.
- In dev mode use `--route-prefix` to mount the app under a sub-path.
