package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/term"
)

func handleLines(args []string) error {
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	var actions []string
	n := len(args)
	i := 0
	for ; i < n; i++ {
		ok := true
		switch args[i] {
		case "sort", "reverse", "uniq":
			actions = append(actions, args[i])
		default:
			ok = false
		}
		if !ok {
			break
		}
	}

	var useArgs bool
	var inputLines []string
	if !isTTY {
		// not tty, try read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			useArgs = true
		} else {
			inputLines = strings.Split(string(data), "\n")
		}
	} else {
		useArgs = true
	}
	if useArgs {
		for _, arg := range args[i:] {
			argLines := strings.Split(arg, "\n")
			inputLines = append(inputLines, argLines...)
		}
	}
	lines := trimSpace(inputLines)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		// trim trailing empty line
		lines = lines[:len(lines)-1]
	}
	for _, action := range actions {
		switch action {
		case "sort":
			lines = sortLines(lines)
		case "reverse":
			lines = reverse(lines)
		case "uniq":
			lines = uniq(lines)
		default:
			return fmt.Errorf("unknown line operation: %s", action)
		}
	}
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}

type SortType int

const (
	SortTypeNone SortType = iota
	SortTypeAsc
	SortTypeDesc
)

func trimSpace(lines []string) []string {
	trimmedLines := make([]string, len(lines))
	for i, line := range lines {
		trimmedLines[i] = strings.TrimSpace(line)
	}
	return trimmedLines
}

func uniq(lines []string) []string {
	mapping := make(map[string]int, len(lines))
	n := len(lines)
	uniqLines := make([]string, 0, len(lines))
	for i := n - 1; i >= 0; i-- {
		line := lines[i]
		if _, ok := mapping[line]; ok {
			continue
		}
		mapping[line] = i
		uniqLines = append(uniqLines, line)
	}
	return reverse(uniqLines)
}

func sortLines(lines []string) []string {
	sortedLines := make([]string, len(lines))
	copy(sortedLines, lines)
	sort.Strings(sortedLines)
	return sortedLines
}

func reverse(lines []string) []string {
	n := len(lines)
	reversedLines := make([]string, n)
	for i, line := range lines {
		reversedLines[n-1-i] = line
	}
	return reversedLines
}
