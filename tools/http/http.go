package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/xhd2015/kool/pkgs/flag"
)

// Handle is the entry point for the HTTP tools
func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("available commands: serve")
	}

	switch args[0] {
	case "serve":
		return handleServe(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// handleServe implements a static file server
func handleServe(args []string) error {
	// Default port
	port := 8080

	// Parse flags
	var remainArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		flag, value := flag.ParseFlag(args, &i)
		if flag == "" {
			remainArgs = append(remainArgs, args[i])
			continue
		}
		switch flag {
		case "--port":
			value, ok := value()
			if !ok {
				return fmt.Errorf("%s requires a port number", flag)
			}
			p, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid port number: %s", value)
			}
			if p < 1 || p > 65535 {
				return fmt.Errorf("port number must be between 1 and 65535")
			}
			port = p
		default:
			return fmt.Errorf("unrecognized: %s", flag)
		}
	}

	if len(remainArgs) > 1 {
		return fmt.Errorf("unexpected argument: %v", remainArgs[1:])
	}

	var dir string
	if len(remainArgs) > 0 {
		dir = remainArgs[0]
	}

	// Create a file server handler
	fs := http.FileServer(http.Dir(dir))

	// Set up the HTTP server
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting HTTP server on http://localhost%s\n", addr)
	fmt.Printf("Serving files from: %s\n", dir)
	fmt.Println("Press Ctrl+C to stop")

	// Start the server
	return http.ListenAndServe(addr, logRequest(fs))
}

// logRequest is a middleware that logs HTTP requests
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}
