# frontend

To install dependencies:

```sh
bun install
```

To run:

```sh
bun run index.ts
```

# build
```sh
bun build ./index.ts --outdir ./build
```

## watch
```sh
bun run watch
``` 

# Server
```go
r.StaticFile("/index.js","frontend/build/index.js")
r.StaticFile("/favicon.ico","frontend/icon.svg")
```

```go
package routehelp

import (
	"github.com/gin-gonic/gin"
)

func StaticFileNoClientSideCache(r *gin.Engine, api string, filePath string) {
	r.GET(api, func(c *gin.Context) {
		// since Chrome/Firefox caches js files aggressively
		// we need to disable fully client-side cache
		c.Header("Cache-Control", "no-cache, must-revalidate")
		c.File(filePath)
	})
}
```