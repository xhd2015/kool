//go:build ignore

package server

import (
	"bytes"
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
	"regexp"
	"strconv"
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

// DefaultVitePort is the fixed Vite port used by script/dev (strictPort).
const DefaultVitePort = 6193

// ServeConfig controls HTTP serve for production and --dev.
type ServeConfig struct {
	Port         int
	Dev          bool
	Route        RouteOptions
	VitePort     int  // proxy target in Dev; 0 → DefaultVitePort when ExternalVite
	ExternalVite bool // Dev: proxy only; do not spawn Vite (air / script/dev)
}

func Serve(port int, dev bool, route RouteOptions) error {
	return ServeWithConfig(ServeConfig{
		Port:  port,
		Dev:   dev,
		Route: route,
	})
}

func ServeWithConfig(cfg ServeConfig) error {
	route := NormalizeRoute(cfg.Route)
	routePrefix := route.Prefix
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      MountRoutePrefix(mux, route),
	}

	if cfg.Dev {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c
			cancel()

			if err := httpServer.Close(); err != nil {
				fmt.Printf("Failed to close server: %v\n", err)
			}
		}()

		vitePort := cfg.VitePort
		if cfg.ExternalVite {
			if vitePort <= 0 {
				vitePort = DefaultVitePort
			}
			fmt.Printf("Dev mode: proxying UI to vite :%d (external)\n", vitePort)
		} else {
			var subProcessDone chan struct{}
			var err error
			vitePort, subProcessDone, err = EnsureFrontendDevServer(ctx, routePrefix)
			if err != nil {
				return err
			}
			if subProcessDone != nil {
				defer func() {
					fmt.Println("Waiting for frontend dev server to be closed...")
					<-subProcessDone
				}()
			}
		}

		if err := ProxyDev(mux, vitePort, route); err != nil {
			return err
		}
	} else {
		err := Static(mux, StaticOptions{Route: route})
		if err != nil {
			return err
		}
	}

	err := RegisterAPI(mux)
	if err != nil {
		return err
	}

	fmt.Printf("Serving directory preview at %s\n", localURL(cfg.Port, route, "/"))
	printRootRoute(cfg.Port, route)

	return httpServer.ListenAndServe()
}

func ProxyDev(mux *http.ServeMux, vitePort int, route RouteOptions) error {
	targetURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", vitePort))
	if err != nil {
		return fmt.Errorf("invalid proxy target: %v", err)
	}
	return proxyDevTarget(mux, targetURL, route)
}

// DevHandler builds the development handler used by an external supervisor.
// It proxies frontend requests to frontendURL, serves API routes locally, and
// mounts the complete app below route.Prefix — plus the root route when
// route.KeepRoot is set, so one process can answer both entry points.
func DevHandler(frontendURL string, route RouteOptions) (http.Handler, error) {
	targetURL, err := url.Parse(frontendURL)
	if err != nil {
		return nil, fmt.Errorf("invalid frontend URL: %v", err)
	}
	mux := http.NewServeMux()
	if err := proxyDevTarget(mux, targetURL, route); err != nil {
		return nil, err
	}
	if err := RegisterAPI(mux); err != nil {
		return nil, err
	}
	return MountRoutePrefix(mux, route), nil
}

func proxyDevTarget(mux *http.ServeMux, targetURL *url.URL, route RouteOptions) error {
	if targetURL.Scheme == "" || targetURL.Host == "" {
		return fmt.Errorf("invalid proxy target: %q", targetURL.String())
	}
	route = NormalizeRoute(route)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	defaultDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		requestPrefix := RequestRoutePrefix(r)
		defaultDirector(r)
		if route.Prefix != "" {
			r.URL.Path = joinRoutePrefix(route.Prefix, r.URL.Path)
			if r.URL.RawPath != "" {
				r.URL.RawPath = joinRoutePrefix(route.Prefix, r.URL.RawPath)
			}
		}
		// The HTML rewrite below needs the prefix the browser actually used,
		// which the inbound request carries in its context.
		*r = *withRequestRoutePrefix(r, requestPrefix)
	}
	// Vite's own HTML is served through this proxy in dev, so the prefix has to
	// be injected per request here too.
	proxy.ModifyResponse = func(resp *http.Response) error {
		if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
			return nil
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		html := prepareFrontendHTML(body, RequestRoutePrefix(resp.Request), true)
		resp.Body = io.NopCloser(bytes.NewReader(html))
		resp.ContentLength = int64(len(html))
		resp.Header.Set("Content-Length", strconv.Itoa(len(html)))
		return nil
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.Host = targetURL.Host
		proxy.ServeHTTP(w, r)
	})
	return nil
}

type StaticOptions struct {
	IndexHtml string // Custom HTML content to serve instead of embedded index.html
	Route     RouteOptions
}

func Static(mux *http.ServeMux, opts StaticOptions) error {
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
			w.Write(prepareFrontendHTML([]byte(opts.IndexHtml), RequestRoutePrefix(r), true))
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

		w.Write(prepareFrontendHTML(content, RequestRoutePrefix(r), true))
	})
	return nil
}

// RouteOptions mounts the app below one optional path prefix.
type RouteOptions struct {
	// Prefix is a path like /demo. Empty or "/" means the host root.
	Prefix string
	// KeepRoot also serves the unprefixed root route, so one process can answer
	// a prefixed front door and a direct domain at the same time.
	KeepRoot bool
}

// NormalizeRoute cleans a route: the prefix takes its canonical form and an
// empty prefix means the root route.
func NormalizeRoute(route RouteOptions) RouteOptions {
	route.Prefix = NormalizeRoutePrefix(route.Prefix)
	return route
}

// NormalizeRoutePrefix converts values like "my-app" and "/my-app/" to
// "/my-app". Empty and "/" mean the app is mounted at the host root.
func NormalizeRoutePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return ""
	}
	prefix = path.Clean("/" + strings.Trim(prefix, "/"))
	if prefix == "/" {
		return ""
	}
	return prefix
}

// ValidateRoutePrefix rejects prefixes that are not plain path segments, so a
// bad flag fails at startup instead of serving broken links.
func ValidateRoutePrefix(prefix string) error {
	normalized := NormalizeRoutePrefix(prefix)
	if normalized == "" {
		return nil
	}
	for _, segment := range strings.Split(strings.TrimPrefix(normalized, "/"), "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("invalid route prefix %q: use path segments like /demo", prefix)
		}
		for _, r := range segment {
			valid := r == '-' || r == '_' || r == '.' ||
				(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
			if !valid {
				return fmt.Errorf("invalid route prefix %q: use path segments like /demo", prefix)
			}
		}
	}
	return nil
}

// routePrefixKey carries the route prefix the current request arrived under.
type routePrefixKey struct{}

// RequestRoutePrefix is the prefix this request arrived under: the mounted
// prefix, or "" when it came in on the root route.
func RequestRoutePrefix(r *http.Request) string {
	if r == nil {
		return ""
	}
	if value, ok := r.Context().Value(routePrefixKey{}).(string); ok {
		return value
	}
	return ""
}

func withRequestRoutePrefix(r *http.Request, prefix string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), routePrefixKey{}, prefix))
}

// MountRoutePrefix mounts handler below route.Prefix, stripping the prefix
// before the request reaches handler. The effective prefix travels in the
// request context so the HTML handler can inject it per request: one process
// then serves both /demo/… (prefix kept) and /… (prefix cleared, KeepRoot).
// An empty prefix leaves handler unchanged.
func MountRoutePrefix(handler http.Handler, route RouteOptions) http.Handler {
	route = NormalizeRoute(route)
	prefix := route.Prefix
	if prefix == "" {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := r.URL.Path
		switch {
		case requestPath == prefix:
			http.Redirect(w, r, prefix+"/", http.StatusTemporaryRedirect)
		case strings.HasPrefix(requestPath, prefix+"/"):
			handler.ServeHTTP(w, withRequestRoutePrefix(stripRoutePrefix(r, prefix), prefix))
		case route.KeepRoot:
			handler.ServeHTTP(w, withRequestRoutePrefix(r, ""))
		default:
			http.NotFound(w, r)
		}
	})
}

func stripRoutePrefix(r *http.Request, prefix string) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	u2 := new(url.URL)
	*u2 = *r.URL
	u2.Path = strings.TrimPrefix(r.URL.Path, prefix)
	if u2.Path == "" {
		u2.Path = "/"
	}
	if u2.RawPath != "" {
		u2.RawPath = strings.TrimPrefix(u2.RawPath, prefix)
		if u2.RawPath == "" {
			u2.RawPath = "/"
		}
	}
	r2.URL = u2
	return r2
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
	if prefixRootPath(requestPath, routePrefix) == requestPath {
		// Already under the prefix (a Vite --base page): never double it.
		return requestPath
	}
	return routePrefix + requestPath
}

// prefixRootPath puts the prefix in front of a root-absolute path once.
func prefixRootPath(pathname, routePrefix string) string {
	if pathname == routePrefix || strings.HasPrefix(pathname, routePrefix+"/") {
		return pathname
	}
	return routePrefix + pathname
}

func localURL(port int, route RouteOptions, requestPath string) string {
	return fmt.Sprintf("http://localhost:%d%s", port, joinRoutePrefix(route.Prefix, requestPath))
}

// printRootRoute reports the second entry point of a dual-route deployment, so
// the root route is visible at startup instead of only in the docs.
func printRootRoute(port int, route RouteOptions) {
	route = NormalizeRoute(route)
	if route.Prefix == "" {
		return
	}
	if route.KeepRoot {
		fmt.Printf("Root route: %s\n", localURL(port, RouteOptions{}, "/"))
		return
	}
	fmt.Printf("Root route: disabled (pass --keep-root-route to serve it)\n")
}

func prepareFrontendHTML(indexHTML []byte, routePrefix string, rewriteRootAssets bool) []byte {
	routePrefix = NormalizeRoutePrefix(routePrefix)
	html := string(indexHTML)
	if rewriteRootAssets {
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

var rootAbsoluteAttrRe = regexp.MustCompile(`(src|href)="(/[^"]*)"`)
var moduleImportRe = regexp.MustCompile(`(import\s+")(/[^"]*)"`)

// prefixRootAbsoluteHTMLAttrs rewrites root-absolute asset URLs so a page served
// under a prefix loads its own files. Values already under the prefix are left
// alone, which keeps the rewrite safe for HTML a dev server already prefixed
// with --base.
func prefixRootAbsoluteHTMLAttrs(html string, routePrefix string) string {
	routePrefix = NormalizeRoutePrefix(routePrefix)
	if routePrefix == "" {
		return html
	}
	html = rootAbsoluteAttrRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := rootAbsoluteAttrRe.FindStringSubmatch(match)
		return parts[1] + `="` + prefixRootPath(parts[2], routePrefix) + `"`
	})
	return moduleImportRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := moduleImportRe.FindStringSubmatch(match)
		return parts[1] + prefixRootPath(parts[2], routePrefix) + `"`
	})
}

func RegisterAPI(mux *http.ServeMux) error {
	// ping
	mux.HandleFunc("/ping", handlePing)

	RegisterCounterAPI(mux)
	RegisterPageMetaAPI(mux)
	RegisterImagesAPI(mux)

	// Unknown API paths get a JSON 404 instead of the SPA fallback, so CLI
	// and agent callers see a clean error. Specific routes above win over
	// this subtree pattern.
	mux.HandleFunc("/api/", handleAPINotFound)

	return nil
}

func handleAPINotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error":"not found: %s"}`+"\n", r.URL.Path)
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
