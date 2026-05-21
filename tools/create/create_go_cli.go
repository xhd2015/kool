package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const goCLIHelp = `
Usage: kool create go-cli <dir>

Create a new Go CLI project.
`

func HandleCreateGoCLI(args []string) error {
	args, err := flags.Help("-h,--help", goCLIHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	projectDir := args[0]
	args = args[1:]
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args, ","))
	}

	_, err = prepareProjectDir(projectDir)
	if err != nil {
		return err
	}

	modulePath, _ := suggestGoModPath(projectDir)
	if modulePath == "" {
		modulePath = filepath.Base(projectDir)
	}

	err = copyTemplateDir(goCLITemplateFS, "go_cli_template", projectDir, filepath.Base(projectDir), modulePath)
	if err != nil {
		return err
	}

	err = os.Rename(filepath.Join(projectDir, "go.mod.template"), filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return fmt.Errorf("failed to rename go.mod.template to go.mod: %v", err)
	}

	err = cmd.Debug().Dir(projectDir).Run("go", "mod", "tidy")
	if err != nil {
		return err
	}

	err = cmd.Debug().Dir(projectDir).Run("git", "init")
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created new go-cli project: %s\n", projectDir)
	return nil
}
