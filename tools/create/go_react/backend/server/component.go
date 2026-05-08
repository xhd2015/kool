package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type ServeOptions struct {
	Static      StaticOptions
	Route       func(mux *http.ServeMux) error // Optional custom route registration
	Dev         bool
	RoutePrefix string
}

func ServeComponent(port int, opts ServeOptions) error {
	if port == 0 {
		var err error
		port, err = FindAvailablePort(8080, 100)
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      mux,
	}

	if opts.Dev {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			cancel()
			server.Close()
		}()

		vitePort, subProcessDone, err := EnsureFrontendDevServer(ctx, opts.RoutePrefix)
		if err != nil {
			return err
		}
		if subProcessDone != nil {
			defer func() {
				fmt.Println("Waiting for frontend dev server to be closed...")
				<-subProcessDone
			}()
		}

		err = ProxyDev(mux, vitePort, opts.RoutePrefix)
		if err != nil {
			return err
		}
	} else {
		staticOpts := opts.Static
		staticOpts.RoutePrefix = opts.RoutePrefix
		err := Static(mux, staticOpts)
		if err != nil {
			return err
		}
	}

	err := RegisterAPI(mux)
	if err != nil {
		return err
	}

	// Register custom routes if provided
	if opts.Route != nil {
		err = opts.Route(mux)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Serving at %s\n", localURL(port, opts.RoutePrefix, "/"))

	server.Handler = mountRoutePrefix(opts.RoutePrefix, mux)
	return server.ListenAndServe()
}

// FormatOptions contains the options for formatting the template HTML
type FormatOptions struct {
	Title          string // __TITLE__ placeholder
	Render         string // __RENDER__ placeholder (default: "renderComponent")
	Component      string // __COMPONENT__ placeholder (mandatory)
	ComponentProps string // __COMPONENT_PROPS__ placeholder (default: "{}")
}

// FormatTemplateHtml formats the template HTML with the given options
func FormatTemplateHtml(opts FormatOptions) (string, error) {
	// Validate mandatory field
	if opts.Component == "" {
		return "", fmt.Errorf("requires component")
	}

	// Set defaults
	title := opts.Title
	if title == "" {
		title = "Untitled"
	}

	render := opts.Render
	if render == "" {
		render = "renderComponent"
	}

	componentProps := opts.ComponentProps
	if componentProps == "" {
		componentProps = "{}"
	}

	// Replace placeholders
	result := templateHTML
	result = strings.ReplaceAll(result, "__TITLE__", title)
	result = strings.ReplaceAll(result, "__RENDER__", render)
	result = strings.ReplaceAll(result, "__COMPONENT__", opts.Component)
	result = strings.ReplaceAll(result, "__COMPONENT_PROPS__", componentProps)

	return result, nil
}
