package create

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const reactHelp = `
kool create react helps to create react based project.

Usage: kool create react [OPTIONS]

Options:
  --dir <dir>                      set the output directory
  --chakra-ui                      using @chakra-ui/typescript
  --help 						   show help message
  -v,--verbose                     show verbose info  

Examples:
  kool create react create my_project            create a new project named my_project
`

func HandleCreateReact(args []string) error {
	if len(args) > 0 && args[0] == "help" {
		fmt.Print(strings.TrimPrefix(reactHelp, "\n"))
		return nil
	}
	n := len(args)
	var chakraUI bool
	var reactScript bool
	var remainArgs []string
	for i := 0; i < n; i++ {
		flag, value := flags.ParseIndex(args, &i)
		if flag == "" {
			remainArgs = append(remainArgs, args[i])
			continue
		}
		_ = value
		switch flag {
		case "--chakra-ui":
			chakraUI = true
		case "--react-script":
			reactScript = true
		case "-h", "--help":
			fmt.Print(strings.TrimPrefix(reactHelp, "\n"))
			return nil
		// ...
		default:
			return fmt.Errorf("unrecognized flag: %s", flag)
		}
	}

	if len(remainArgs) == 0 {
		return fmt.Errorf("requries project_name")
	}
	if len(remainArgs) > 1 {
		return fmt.Errorf("requires only one project_name, given: %s", strings.Join(remainArgs[1:], ","))
	}
	projectName := remainArgs[0]
	createArgs := remainArgs[1:]

	// npm create vite@latest my-app -- --template react-ts

	// npx create-react-app my-app
	rcTemplate := "react-ts"
	scriptName := "vite@latest"
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
	runArgs = append(runArgs, "--template", rcTemplate)
	runArgs = append(runArgs, createArgs...)
	err := cmd.Debug().Run("npx", runArgs...)
	if err != nil {
		return err
	}
	return nil
}
