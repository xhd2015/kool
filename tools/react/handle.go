package react

import "fmt"

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires sub commands: example")
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "example":
		return handleExample(args)
	}

	return fmt.Errorf("unknown command: %s", cmd)
}
