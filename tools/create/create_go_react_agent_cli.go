package create

import (
	"embed"
)

//go:embed all:go_react_agent_cli
var goReactAgentCLITemplateFS embed.FS

const goReactAgentCLIHelp = `
Usage: kool create go-react-agent-cli [--go-module <module-path>] <project-name>

Create a new go-react project with an agent-driven CLI: the binary starts the
web UI and API server, and get/put/post/delete sub-commands talk to the
running server so agents (and humans) can drive the app from a terminal.
Ships a persistent counter demo and an embedded agent skill.

  --go-module  specify the Go module path (e.g. github.com/user/repo)
               otherwise auto-detected from git remote, falls back to <project-name>
`

func HandleCreateGoReactAgentCLI(args []string) error {
	return createGoReactProject(goReactAgentCLITemplateFS, "go_react_agent_cli", "go-react-agent-cli", goReactAgentCLIHelp, args)
}
