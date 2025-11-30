package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xhd2015/kool/pkgs/web"
)

type ServeOptions struct {
	Static         StaticOptions
	NoOpenBrowser  bool
	OpenBrowserUrl func(port int, url string) string
	Route          func(mux *http.ServeMux) error // Optional custom route registration
	Dev            bool
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
		err := ProxyDev(mux)
		if err != nil {
			return err
		}
	} else {
		err := Static(mux, opts.Static)
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

	url := fmt.Sprintf("http://localhost:%d", port)

	fmt.Printf("Serving at %s\n", url)

	if !opts.NoOpenBrowser {
		go func() {
			time.Sleep(1 * time.Second)
			openUrl := url
			if opts.OpenBrowserUrl != nil {
				openUrl = opts.OpenBrowserUrl(port, url)
			}
			web.OpenBrowser(openUrl)
		}()
	}

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
