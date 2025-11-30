package main

import (
	"embed"
	"fmt"
	"os"

	"MODULE_NAME/run"
	"MODULE_NAME/server"
)

//go:embed PROJECT_NAME-react/dist
var distFS embed.FS

func main() {
	server.Init(distFS)

	err := run.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
