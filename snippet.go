package main

import (
	_ "embed"
	"fmt"
)

//go:embed snippets/go.cli.template
var goCliTemplate string

type Snippet struct {
	Name    string
	Content string
}

var Snippets = []Snippet{
	{
		Name:    "go.cli",
		Content: goCliTemplate,
	},
}

func handleSnippet(args []string) error {
	if len(args) == 0 {
		fmt.Println("Available snippets:")
		for _, snippet := range Snippets {
			fmt.Printf("  %s\n", snippet.Name)
		}
		return nil
	}

	for _, snippet := range Snippets {
		if snippet.Name == args[0] {
			fmt.Println(snippet.Content)
			return nil
		}
	}

	return fmt.Errorf("snippet %s not found", args[0])
}
