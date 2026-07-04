package main

import (
	"embed"
	"fmt"
	"os"

	"__MODULE_NAME__/run"
	"__MODULE_NAME__/server"
)

//go:embed __PROJECT_NAME__-react/dist
var distFS embed.FS

//go:embed __PROJECT_NAME__-react/template.html
var templateHTML string

func main() {
	server.Init(distFS, templateHTML)

	err := run.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
