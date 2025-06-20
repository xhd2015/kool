package example

import (
	"fmt"
	"strings"

	_ "embed"
)

//go:embed parse_flag_template.go
var parseFlagTemplate string

//go:embed debug_template.txt
var debugTemplate string

func Handle(args []string, legacyFlagSnippet string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kool go example <snippet>\navailable snippets: parse-flag, debug")
	}
	snippet := args[0]
	args = args[1:]
	switch snippet {
	case "parse-flag-legacy":
		fmt.Print(legacyFlagSnippet)
	case "parse-flag":
		var cliName string
		if len(args) > 0 {
			cliName = args[0]
			args = args[1:]
			if len(args) > 0 {
				return fmt.Errorf("unrecognized extra arguments: %v", args)
			}
		}

		code := parseFlagTemplate
		if idx := strings.Index(parseFlagTemplate, "import ("); idx >= 0 {
			code = parseFlagTemplate[idx:]
		}
		code = strings.ReplaceAll(code, "cli", cliName)
		fmt.Print(strings.ReplaceAll(code, "\t", "  "))
	case "debug":
		fmt.Print(debugTemplate)
	}
	return nil
}
