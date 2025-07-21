package create

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const reactHelp = `
kool create react helps to create react based project.

Usage: kool create react [OPTIONS] <project_name>

Options:
  --dir <dir>                      set the output directory
  --chakra-ui                      using @chakra-ui/typescript
  --bun                            use bun instead of npx (auto use bun if possible)
  --react-script                   use create-react-app instead of vite
  --help 						   show help message
  -v,--verbose                     show verbose info  

Examples:
  kool create react create my_project            create a new project named my_project

References:
  https://vite.dev/guide/
`

func HandleCreateReact(args []string) error {
	var chakraUI bool
	var reactScript bool
	var bun *bool // auto
	args, err := flags.Bool("--chakra-ui", &chakraUI).
		Bool("--react-script", &reactScript).
		Bool("--bun", &bun).
		Help("-h,--help", reactHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires projectName, try `kool create react --help`")
	}

	projectName := args[0]
	createArgs := args[1:]

	if projectName == "--help" || projectName == "help" {
		fmt.Print(strings.TrimPrefix(reactHelp, "\n"))
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("requires only one project_name, given: %s", strings.Join(args, ","))
	}

	var useBun bool
	engine := "npx"
	if bun == nil || *bun {
		// select bunx if possible
		_, lookupErr := exec.LookPath("bun")
		if lookupErr != nil {
			if *bun {
				return fmt.Errorf("bun is not installed, please install it first")
			}
		} else {
			engine = "bun"
			useBun = true
		}
	}
	//  bun create vite my-vue-app --template react-ts
	if useBun && !reactScript {
		if chakraUI {
			return fmt.Errorf("chakra-ui is only supported with create-react-app")
		}
		err := cmd.Debug().Run(engine, "create", "vite", projectName, "--template", "react-ts")
		if err != nil {
			return err
		}
		err = cmd.Debug().Dir(projectName).Run("bun", "install")
		if err != nil {
			return err
		}
		return nil
	}

	// npm create vite@latest my-app -- --template react-ts

	// npx create-react-app my-app
	scriptName := "vite@latest"
	rcTemplate := "react-ts"
	if reactScript {
		scriptName = "create-react-app"
		rcTemplate = "typescript"
	}
	runArgs := []string{"-y", scriptName}
	runArgs = append(runArgs, projectName)
	if chakraUI {
		if !reactScript {
			return fmt.Errorf("chakra-ui is only supported with create-react-app")
		}
		rcTemplate = "@chakra-ui/typescript"
	}
	if !reactScript {
		runArgs = append(runArgs, "--")
	}
	runArgs = append(runArgs, "--template", rcTemplate)
	runArgs = append(runArgs, createArgs...)

	return cmd.Debug().Run(engine, runArgs...)
}
