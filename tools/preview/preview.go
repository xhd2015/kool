package preview

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/kool/tools/preview/viewer"
	"github.com/xhd2015/less-gen/flags"
)

const help = `

Options:
  --plant-uml-server ADDR    plantuml server url, default is https://www.plantuml.com/plantuml, can be http://localhost:8080

Example plantuml server:
  docker run --rm -p 8080:8080 plantuml/plantuml-server:jetty
`

// TODO:
// - [ ] avoid previewing binary files, just like vscode
// - [ ] use websocket to sync the backend and frontend content change
// - [ ] only show save retry button when save failed, no other status needed
// - [ ] split the html into multiple files and components
// - [ ] use user's default shell
// - [x] terminal line wrap working test
// - [x] allow edit arbitrary txt files
// - [x] markdown: open link in new tab
// - [x] avoid overriding inconsistent file
// - [x] fix resize issue
// - [x] add a use `docker` option to start plantuml server in docker
// - [ ] mermaid preview react support
// - [ ] remember per-file zoom state
func Handle(args []string) error {
	var plantumlServer string

	args, err := flags.String("--plant-uml-server", &plantumlServer).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires a file or directory to preview")
	}

	path := args[0]

	// Check if path exists
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// If it's a directory, use the viewer
	if stat.IsDir() {
		return viewer.Serve(path, plantumlServer)
	}
	// Use the viewer for UML files (it has built-in UML support)
	// Get absolute path for the initial file
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	// Use current working directory as root
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}
	return viewer.ServeWithInitialFile(cwd, plantumlServer, absPath)
}
