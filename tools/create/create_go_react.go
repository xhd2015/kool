package create

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
)

//go:embed go_react/server.go
var goReactServer string

const goReactHelp = `	
Usage: kool create go-react <project-name>

Create a new go-react project.
`

func HandleCreateGoReact(args []string) error {
	args, err := flags.Help("-h,--help", goReactHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires project")
	}
	projectName := args[0]
	args = args[1:]

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args, ","))
	}

	if _, statErr := os.Stat(projectName); statErr == nil {
		return fmt.Errorf("project %s already exists", projectName)
	}

	err = os.MkdirAll(projectName, 0755)
	if err != nil {
		return err
	}

	engine := "bun"

	reactProjectName := projectName + "-react"
	err = cmd.Debug().Dir(projectName).Run(engine, "create", "vite", reactProjectName, "--template", "react-ts")
	if err != nil {
		return err
	}
	reactDir := filepath.Join(projectName, reactProjectName)
	err = cmd.Debug().Dir(reactDir).Run("bun", "install")
	if err != nil {
		return err
	}
	err = cmd.Debug().Dir(reactDir).Run("bun", "run", "build")
	if err != nil {
		return err
	}

	err = os.Rename(filepath.Join(reactDir, "public", "vite.svg"), filepath.Join(reactDir, "public", projectName+".svg"))
	if err != nil {
		return err
	}

	err = replaceFile(filepath.Join(reactDir, "index.html"), "/vite.svg", "/"+projectName+".svg")
	if err != nil {
		return err
	}

	goReactServerFile := filepath.Join(projectName, "server.go")

	goReactServer = strings.Replace(goReactServer, "package server", "package "+projectName, 1)
	goReactServer = strings.ReplaceAll(goReactServer, "react/dist", reactProjectName+"/dist")

	err = os.WriteFile(goReactServerFile, []byte(goReactServer), 0644)
	if err != nil {
		return err
	}

	return nil
}

func replaceFile(f string, old, new string) error {
	content, err := os.ReadFile(f)
	if err != nil {
		return err
	}
	content = []byte(strings.Replace(string(content), old, new, 1))
	return os.WriteFile(f, content, 0644)
}
