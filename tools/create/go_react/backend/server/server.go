package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var distFS embed.FS
var templateHTML string

func Init(fs embed.FS, tmpl string) {
	distFS = fs
	templateHTML = tmpl
}

func checkPort(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// EnsureFrontendDevServer starts `bun run dev` for the React app on an
// auto-selected free port (starting at 5173) and blocks until the port
// is reachable. It returns the chosen port along with a channel that
// closes once the sub-process has fully terminated after ctx is
// cancelled.
//
// Selecting a free port dynamically means multiple projects can run in
// --dev mode in parallel; it also avoids accidentally proxying to
// another project's vite instance that happens to own 5173.
func EnsureFrontendDevServer(ctx context.Context, routePrefix string) (int, chan struct{}, error) {
	vitePort, err := findFreeVitePort(5173, 100)
	if err != nil {
		return 0, nil, fmt.Errorf("pick frontend port: %v", err)
	}

	fmt.Printf("Starting frontend dev server on port %d...\n", vitePort)
	// `--` separates bun-run flags from the script's own flags so vite
	// picks up `--port`.
	cmdArgs := []string{"run", "dev", "--", "--port", fmt.Sprintf("%d", vitePort)}
	if base := NormalizeRoutePrefix(routePrefix); base != "" {
		cmdArgs = append(cmdArgs, "--base", strings.TrimRight(base, "/")+"/")
	}
	cmd := exec.Command("bun", cmdArgs...)
	cmd.Dir = "__PROJECT_NAME__-react/"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to start frontend dev server: %v", err)
	}

	// childExited closes once the sub-process is fully reaped.
	childExited := make(chan struct{})
	go func() {
		cmd.Wait()
		close(childExited)
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				fmt.Println("Stopping frontend dev server...")
				cmd.Process.Kill()
			}
			<-childExited
		case <-childExited:
		}
	}()

	fmt.Printf("Waiting for frontend server on port %d...", vitePort)
	for i := 0; i < 30; i++ {
		// Exit the ready loop immediately if vite died (e.g. port in
		// use) so we surface the failure instead of hanging 30s.
		select {
		case <-childExited:
			fmt.Println()
			return 0, nil, fmt.Errorf("frontend dev server exited before it became ready")
		default:
		}
		if checkPort(vitePort) {
			fmt.Println(" Ready!")
			return vitePort, done, nil
		}
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()
	return 0, nil, fmt.Errorf("frontend server failed to start within timeout")
}

// findFreeVitePort returns the first port >= startPort that has
// nothing listening on localhost. Unlike FindAvailablePort (which uses
// net.Listen on :port and can succeed even when another process has
// bound the loopback interface on the same port family), this uses
// `checkPort` so the result reflects "can vite's default loopback
// listener use this port?"
func findFreeVitePort(startPort, maxAttempts int) (int, error) {
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		if !checkPort(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port found in [%d, %d)", startPort, startPort+maxAttempts)
}

func Serve(port int, dev bool, routePrefix string) error {
	routePrefix = NormalizeRoutePrefix(routePrefix)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      mountRoutePrefix(routePrefix, mux),
	}

	if dev {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			cancel()

			if err := server.Close(); err != nil {
				fmt.Printf("Failed to close server: %v\n", err)
			}
		}()

		vitePort, subProcessDone, err := EnsureFrontendDevServer(ctx, routePrefix)
		if err != nil {
			return err
		}
		if subProcessDone != nil {
			defer func() {
				fmt.Println("Waiting for frontend dev server to be closed...")
				<-subProcessDone
			}()
		}

		err = ProxyDev(mux, vitePort, routePrefix)
		if err != nil {
			return err
		}
	} else {
		err := Static(mux, StaticOptions{RoutePrefix: routePrefix})
		if err != nil {
			return err
		}
	}

	err := RegisterAPI(mux)
	if err != nil {
		return err
	}

	fmt.Printf("Serving directory preview at %s\n", localURL(port, routePrefix, "/"))

	return server.ListenAndServe()
}

func ProxyDev(mux *http.ServeMux, vitePort int, routePrefix string) error {
	targetURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", vitePort))
	if err != nil {
		return fmt.Errorf("invalid proxy target: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	routePrefix = NormalizeRoutePrefix(routePrefix)
	defaultDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		defaultDirector(r)
		if routePrefix == "" {
			return
		}
		r.URL.Path = joinRoutePrefix(routePrefix, r.URL.Path)
		if r.URL.RawPath != "" {
			r.URL.RawPath = joinRoutePrefix(routePrefix, r.URL.RawPath)
		}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.Host = targetURL.Host
		proxy.ServeHTTP(w, r)
	})
	return nil
}

type StaticOptions struct {
	IndexHtml   string // Custom HTML content to serve instead of embedded index.html
	RoutePrefix string
}

func Static(mux *http.ServeMux, opts StaticOptions) error {
	routePrefix := NormalizeRoutePrefix(opts.RoutePrefix)
	// Serve static files from the embedded React build
	reactFileSystem, err := fs.Sub(distFS, "__PROJECT_NAME__-react/dist")
	if err != nil {
		return fmt.Errorf("failed to create react file system: %v", err)
	}

	// Create sub-filesystem for assets
	assetsFileSystem, err := fs.Sub(reactFileSystem, "assets")
	if err != nil {
		return fmt.Errorf("failed to create assets file system: %v", err)
	}

	// Serve React assets from /assets/ path with proper MIME types

	// Serve index.css and index.js from assets with pattern matching
	mux.HandleFunc("/assets/index.css", func(w http.ResponseWriter, r *http.Request) {
		serveAssetWithPattern(w, r, assetsFileSystem, "index.css", "index-", ".css", "text/css")
	})
	mux.HandleFunc("/assets/index.js", func(w http.ResponseWriter, r *http.Request) {
		serveAssetWithPattern(w, r, assetsFileSystem, "index.js", "index-", ".js", "application/javascript")
	})

	mux.Handle("/assets/", http.StripPrefix("/assets/", &mimeTypeHandler{http.FileServer(http.FS(assetsFileSystem))}))
	// Serve React static files like vite.svg from root
	mux.Handle("/__PROJECT_NAME__.svg", &mimeTypeHandler{http.FileServer(http.FS(reactFileSystem))})

	// Serve the main HTML page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		// Use custom IndexHtml if provided
		if opts.IndexHtml != "" {
			w.Write(prepareFrontendHTML([]byte(opts.IndexHtml), routePrefix, true))
			return
		}

		// Otherwise, serve embedded index.html
		indexFile, err := reactFileSystem.Open("index.html")
		if err != nil {
			http.Error(w, "Failed to load index.html", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		content, err := io.ReadAll(indexFile)
		if err != nil {
			http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
			return
		}

		w.Write(prepareFrontendHTML(content, routePrefix, true))
	})
	return nil
}

// NormalizeRoutePrefix converts values like "my-app" and "/my-app/" to
// "/my-app". Empty and "/" mean the app is mounted at the host root.
func NormalizeRoutePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return ""
	}
	return path.Clean("/" + strings.Trim(prefix, "/"))
}

func mountRoutePrefix(routePrefix string, handler http.Handler) http.Handler {
	routePrefix = NormalizeRoutePrefix(routePrefix)
	if routePrefix == "" {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == routePrefix {
			http.Redirect(w, r, routePrefix+"/", http.StatusTemporaryRedirect)
			return
		}
		if !strings.HasPrefix(r.URL.Path, routePrefix+"/") {
			http.NotFound(w, r)
			return
		}

		r2 := new(http.Request)
		*r2 = *r
		u2 := new(url.URL)
		*u2 = *r.URL
		u2.Path = strings.TrimPrefix(r.URL.Path, routePrefix)
		if u2.Path == "" {
			u2.Path = "/"
		}
		if u2.RawPath != "" {
			u2.RawPath = strings.TrimPrefix(u2.RawPath, routePrefix)
			if u2.RawPath == "" {
				u2.RawPath = "/"
			}
		}
		r2.URL = u2
		handler.ServeHTTP(w, r2)
	})
}

func joinRoutePrefix(routePrefix string, requestPath string) string {
	routePrefix = NormalizeRoutePrefix(routePrefix)
	if requestPath == "" {
		requestPath = "/"
	}
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}
	if routePrefix == "" {
		return requestPath
	}
	if requestPath == "/" {
		return routePrefix + "/"
	}
	return routePrefix + requestPath
}

func localURL(port int, routePrefix string, requestPath string) string {
	return fmt.Sprintf("http://localhost:%d%s", port, joinRoutePrefix(routePrefix, requestPath))
}

func prepareFrontendHTML(indexHTML []byte, routePrefix string, rewriteRootAssets bool) []byte {
	routePrefix = NormalizeRoutePrefix(routePrefix)
	html := string(indexHTML)
	if rewriteRootAssets && routePrefix != "" {
		html = prefixRootAbsoluteHTMLAttrs(html, routePrefix)
	}
	routePrefixJSON, _ := json.Marshal(routePrefix)
	script := "<script>window.__KOOL_ROUTE_PREFIX__=" + string(routePrefixJSON) + ";</script>"
	if !strings.Contains(html, "window.__KOOL_ROUTE_PREFIX__") {
		if strings.Contains(html, "</head>") {
			html = strings.Replace(html, "</head>", "  "+script+"\n</head>", 1)
		} else {
			html = script + "\n" + html
		}
	}
	return []byte(html)
}

func prefixRootAbsoluteHTMLAttrs(html string, routePrefix string) string {
	for _, attr := range []string{"src", "href"} {
		needle := attr + `="/`
		replacement := attr + `="` + routePrefix + `/`
		html = strings.ReplaceAll(html, needle, replacement)
	}
	html = strings.ReplaceAll(html, `import "/`, `import "`+routePrefix+`/`)
	return html
}

func RegisterAPI(mux *http.ServeMux) error {
	// ping
	mux.HandleFunc("/ping", handlePing)

	return nil
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

// mimeTypeHandler wraps an http.Handler and sets proper MIME types
type mimeTypeHandler struct {
	handler http.Handler
}

func (h *mimeTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set MIME type based on file extension
	ext := filepath.Ext(r.URL.Path)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		// Use Go's built-in MIME type detection for other files
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
	}

	// Call the wrapped handler
	h.handler.ServeHTTP(w, r)
}

// serveAssetWithPattern finds and serves the first available file matching the given exact match or prefix and suffix
func serveAssetWithPattern(w http.ResponseWriter, r *http.Request, assetsFS fs.FS, exactMatch, prefix, suffix, contentType string) {
	// First try exact match
	if _, err := fs.Stat(assetsFS, exactMatch); err == nil {
		serveAssetFile(w, r, assetsFS, exactMatch, contentType)
		return
	}

	// Then try pattern matching with prefix and suffix
	entries, err := fs.ReadDir(assetsFS, ".")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), suffix) {
			serveAssetFile(w, r, assetsFS, entry.Name(), contentType)
			return
		}
	}

	// No matching file found
	http.NotFound(w, r)
}

// serveAssetFile serves a specific file from the assets filesystem
func serveAssetFile(w http.ResponseWriter, r *http.Request, assetsFS fs.FS, filename string, contentType string) {
	file, err := assetsFS.Open(filename)
	if err != nil {
		http.Error(w, "Failed to open asset file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read asset file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}

// checkPortAvailable checks if a port is available
func checkPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// FindAvailablePort finds a port starting from startPort
func FindAvailablePort(startPort int, maxAttempts int) (int, error) {
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		if checkPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found")
}
