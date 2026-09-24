---
name: __PROJECT_NAME__
description: Agent-driven web app - a Go API server plus a React UI. Drive the running __PROJECT_NAME__ server from the CLI - start it, then get/put/post/delete its API to inspect and change data, upload images into the shared library, and read the ids every entity draws from. Use when the user asks __PROJECT_NAME__ to show, plan, or illustrate something, or mentions the __PROJECT_NAME__ server, its API, its images, or its CLI.
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
| GET    | /api/ids         | the last allocated id (`{"last":N}`)           |
| POST   | /api/ids         | allocate ids (`{"count":n}` → `{"ids":[…]}`)    |
| POST   | /api/images      | upload an image (multipart `file`, `name`)     |
| GET    | /api/images      | audit the library (`?verify=0` skips the byte check) |
| GET    | /api/images/<id> | one image, with its `path` and `usages`        |
| DELETE | /api/images/<id> | delete; refused while a page shows it (`?force=1` detaches first) |
| GET    | /api/data/images/<id>/image.<ext> | the stored bytes            |
| GET    | /api/gallery     | the demo container that references images      |
| PUT    | /api/gallery     | replace it (`{"image_ids":[…]}`)               |
| GET    | /api/page-meta   | every card's server-owned meta (title/hint/empty) |
| GET    | /api/pages/home  | the Home page document: cards in web order, each with `meta`, `empty` and its rows |

Card meta is server-owned: a page document carries `sections[].meta`
(`title`, `hint`, `empty`) next to the rows, so `get /api/pages/home` tells you
what every card is for — empty cards included — without reading the React
source. Author the words in `server/pagemeta/parts/<Card>.json`; the React card
imports that same file over the `@pagemeta/` alias, and
`go test ./server/pagemeta/` fails when a registered card is missing a word.

Ids come from one sequence: `~/.<project>/id.json`, shared by the counter, the
image library and every entity you add. Grow the API on top of this pattern:
one JSON handler per route, registered in `server.RegisterAPI`.

## The image library

An uploaded image is stored once, under `~/.<project>/images/<id>/` — the bytes
as `image.<ext>`, plus the `meta.json` that commits the record. Identical bytes
reuse the existing record, so an id's bytes never change and its URL is
cacheable forever. `server/images*.go` owns the library (store, routes, audit,
delete guard, sniffer) and `server/gallery.go` the demo container and the
container registry behind the delete guard.

```sh
__PROJECT_NAME__ post /api/images --file chengdu.jpg --name 成都
# uploaded image 7 · 成都
# path /Users/you/.__PROJECT_NAME__/images/7/image.jpg
# url  /api/data/images/7/image.jpg

__PROJECT_NAME__ get /api/images          # every image: size, where it is used, state
__PROJECT_NAME__ get /api/images/7        # path first, then name/format/state
__PROJECT_NAME__ delete /api/images/7     # refused: the refusal names the page
__PROJECT_NAME__ delete /api/images/7 --force   # detaches the reference, then deletes
```

- The library is **audited, not trusted**: a file whose bytes are not a picture
  is reported `NOT AN IMAGE (detected text/xml)` even though it served a 200.
  The verdict comes from the bytes; the file name only explains a rejection.
- Deleting an image a container still references is **refused** until the
  reference is detached (`--force` does it). One registry in `server/gallery.go`
  drives both the audit and the detach, so the two can never disagree.
- To add your own container, give it an `image_ids` field and register it in
  `imageContainers`. A file with image references that is *not* registered makes
  the delete refuse rather than leave a dangling id behind.
- `get /api/images/<id>` prints the **absolute path of the stored bytes** on the
  first line, so the next step can open the file. Never inline binary content.

Full recipe: `go-best-practice skill --show storage/unified-assets`, and
`storage/id-allocator` for the shared sequence.

## Walkthrough

```sh
__PROJECT_NAME__ server &                     # start (or: go run ./cmd/__PROJECT_NAME__ server)
__PROJECT_NAME__ get /ping                    # pong
__PROJECT_NAME__ post /api/counter            # allocate id 1 (silent)
__PROJECT_NAME__ get /api/counter --json      # {"last":1}
__PROJECT_NAME__ post /api/images --file photo.jpg --name 成都   # id 2
__PROJECT_NAME__ get /api/images              # the library, audited
__PROJECT_NAME__ get /api/images/2            # the path to open
__PROJECT_NAME__ get /api/pages/home --json   # cards with meta + rows + empty flags
__PROJECT_NAME__ get http://localhost:8080/api/counter   # full-URL form
```

## Tips

- Check the server is up with `__PROJECT_NAME__ get /ping` before other calls.
- Use `--dry-run` to show an agent's intended write before committing to it.
- In dev mode use `--route-prefix` to mount the app under a sub-path; add
  `--keep-root-route` to serve the unprefixed root route at the same time.
- Images are addresses, never black holes: `get /api/images/<id>` prints a path
  a local step can open.
