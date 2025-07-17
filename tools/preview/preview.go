package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/pkgs/web"
	"github.com/xhd2015/kool/tools/preview/uml"
	"github.com/xhd2015/kool/tools/preview/viewer"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const help = `

Options:
  --auto-docker              auto start plantuml server in docker
  --plant-uml-server ADDR    plantuml server url, default is https://www.plantuml.com/plantuml, can be http://localhost:8080

Example plantuml server:
  docker run --rm -p 8080:8080 plantuml/plantuml-server:jetty
`

func Handle(args []string) error {
	var autoDocker bool
	var plantumlServer string

	args, err := flags.String("--plant-uml-server", &plantumlServer).
		Bool("--auto-docker", &autoDocker).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires a file or directory to preview")
	}

	if autoDocker {
		if plantumlServer == "" {
			port, err := web.FindAvailablePort(8080, 100)
			if err != nil {
				return err
			}
			fmt.Printf("Starting plantuml server in docker on port %d\n", port)
			go func() {
				err := cmd.Debug().Run("docker", "run", "--rm", "-p", fmt.Sprintf("%d:7070", port), "plantuml/plantuml-server:jetty")
				if err != nil {
					fmt.Printf("Failed to start plantuml server in docker: %v\n", err)
				}
			}()
			plantumlServer = fmt.Sprintf("http://localhost:%d", port)
		}
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

	// If it's a file, handle based on extension
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".uml" || ext == ".puml" {
		return uml.Serve(path, plantumlServer)
	}

	if ext == ".mmd" || ext == ".md" {
		// For Mermaid and Markdown files, use the viewer (directory viewer can handle individual files)
		return viewer.Serve(filepath.Dir(path), plantumlServer)
	}

	return fmt.Errorf("unsupported file type: %s", ext)
}
