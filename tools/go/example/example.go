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

type Example struct {
	Name   string
	Handle func(args []string) error
}

var examples = []Example{
	{
		Name: "parse-flag",
		Handle: func(args []string) error {
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
			return nil
		},
	},
	{
		Name: "debug",
		Handle: func(args []string) error {
			fmt.Print(debugTemplate)
			return nil
		},
	},
	{
		Name:   "option-pattern",
		Handle: handleOptionPattern,
	},
}

func Handle(args []string) error {
	var names []string
	for _, example := range examples {
		names = append(names, example.Name)
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: kool go example <snippet>\navailable snippets: %s", strings.Join(names, ", "))
	}
	snippet := args[0]
	args = args[1:]

	var example Example
	var found bool
	for _, e := range examples {
		if e.Name == snippet {
			example = e
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unknown snippet: %s, available snippets: %s", snippet, strings.Join(names, ", "))
	}

	return example.Handle(args)
}

func handleOptionPattern(args []string) error {
	var funcName string
	if len(args) > 0 {
		funcName = args[0]
		args = args[1:]
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra arguments: %v", args)
		}
	}

	tpl := `
// Option represents an option for __NAME__
type Option func(cfg *__CONFIG__)

// __CONFIG__ holds configuration for parsing
type __CONFIG__ struct {
    maxDepth    int
    hasMaxDepth bool
}

// WithMaxDepth sets the maximum depth for parsing
func WithMaxDepth(depth int) Option {
    return func(cfg *__CONFIG__) {
        cfg.maxDepth = depth
        cfg.hasMaxDepth = true
    }
}

func __NAME__(rootDir SchemaDir, opts ...Option) (string, error) {
    cfg := &__CONFIG__{}
    for _, opt := range opts {
        opt(cfg)
    }

    doSomething(cfg)
}
`

	if funcName == "" {
		funcName = "DoSomething"
	}

	tpl = strings.ReplaceAll(tpl, "__NAME__", funcName)
	tpl = strings.ReplaceAll(tpl, "__CONFIG__", lowFirst(funcName)+"Config")
	fmt.Print(strings.TrimPrefix(tpl, "\n"))
	return nil
}

func lowFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
