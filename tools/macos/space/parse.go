package space

import (
	"fmt"
	"strconv"
	"strings"
)

func splitRun(args []string) (left, right []string, hasRun bool) {
	for i, a := range args {
		if a == "--run" {
			left = append([]string(nil), args[:i]...)
			if i+1 < len(args) {
				right = append([]string(nil), args[i+1:]...)
			}
			return left, right, true
		}
	}
	return append([]string(nil), args...), nil, false
}

func parseDesktopNumber(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("missing desktop number")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid desktop number: %s", s)
	}
	if n < 1 {
		return 0, fmt.Errorf("desktop number must be >= 1, got %d", n)
	}
	return n, nil
}

func isHelpToken(s string) bool {
	return s == "-h" || s == "--help" || s == "help"
}
