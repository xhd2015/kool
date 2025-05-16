package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, arg)
			continue
		}
		if arg == "--port" {
			if i+1 >= n {
				return fmt.Errorf("--port requires a port number")
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid port number: %s", args[i+1])
			}
			if p < 1 || p > 65535 {
				return fmt.Errorf("port number must be between 1 and 65535")
			}
			port = p
			i++ // Skip the next argument
			continue
		} else if strings.HasPrefix(arg, "--port=") {
			portStr := strings.TrimPrefix(arg, "--port=")
			p, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("invalid port number: %s", portStr)
			}
			if p < 1 || p > 65535 {
				return fmt.Errorf("port number must be between 1 and 65535")
			}
			port = p
			continue
		} else {
			return fmt.Errorf("unrecognized option: %s", arg)
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
