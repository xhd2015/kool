package viewer

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/kool/pkgs/web"
)

//go:embed index.html
var indexHtml string

//go:embed index.css
var indexCss string

//go:embed dark.css
var darkCss string

//go:embed index.js
var indexJs string

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

var plantUMLCacheDir string

func init() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to temp directory if user cache dir is not available
		plantUMLCacheDir = filepath.Join(os.TempDir(), "plant-uml", "svg")
	} else {
		plantUMLCacheDir = filepath.Join(userCacheDir, "plant-uml", "svg")
	}
}

func Serve(dir string) error {
	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(plantUMLCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	port, err := web.FindAvailablePort(8080, 100)
	if err != nil {
		return err
	}

	htmlContent := strings.ReplaceAll(indexHtml, "__CSS__", "<style>\n"+indexCss+"\n"+darkCss+"\n</style>\n")
	htmlContent = strings.ReplaceAll(htmlContent, "__SCRIPT__", "<script>\n"+indexJs+"\n</script>\n")
	// Serve the main HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	})

	// API to get directory tree
	http.HandleFunc("/api/tree", func(w http.ResponseWriter, r *http.Request) {
		tree, err := buildFileTree(absDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tree)
	})

	// API to serve cached PlantUML SVG images
	http.HandleFunc("/planuml/svg/", func(w http.ResponseWriter, r *http.Request) {
		// Extract encoded string from URL path
		path := strings.TrimPrefix(r.URL.Path, "/planuml/svg/")
		if path == "" {
			http.Error(w, "encoded parameter is required", http.StatusBadRequest)
			return
		}

		// Generate cache file path using SHA256 hash of the encoded string for safety
		hash := sha256.Sum256([]byte(path))
		cacheFileName := hex.EncodeToString(hash[:]) + ".svg"
		cacheFilePath := filepath.Join(plantUMLCacheDir, cacheFileName)

		// Check if file exists in cache
		if _, err := os.Stat(cacheFilePath); err == nil {
			// Serve from cache
			http.ServeFile(w, r, cacheFilePath)
			return
		}

		// File not in cache, fetch from PlantUML service
		plantUMLURL := fmt.Sprintf("https://www.plantuml.com/plantuml/svg/%s", path)

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Get(plantUMLURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch from PlantUML: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("PlantUML service returned status: %d", resp.StatusCode), http.StatusBadGateway)
			return
		}

		// Read the response body
		svgData, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read PlantUML response: %v", err), http.StatusInternalServerError)
			return
		}

		// Only cache if we have valid SVG content (basic validation)
		if len(svgData) > 0 && strings.Contains(string(svgData), "<svg") {
			// Save to cache
			if err := os.WriteFile(cacheFilePath, svgData, 0644); err != nil {
				// Log error but don't fail the request
				fmt.Printf("Warning: Failed to cache PlantUML SVG: %v\n", err)
			}
		}

		// Serve the SVG content
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
		w.Write(svgData)
	})

	// API to get file content for preview
	http.HandleFunc("/api/preview", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "path parameter is required", http.StatusBadRequest)
			return
		}

		// Security check: ensure the file is within the served directory
		absFilePath, err := filepath.Abs(filePath)
		if err != nil {
			http.Error(w, "invalid file path", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(absFilePath, absDir) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}

		// Check if file exists
		stat, err := os.Stat(absFilePath)
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		if stat.IsDir() {
			http.Error(w, "cannot preview directory", http.StatusBadRequest)
			return
		}

		ext := strings.ToLower(filepath.Ext(absFilePath))

		// Handle UML files
		if ext == ".uml" || ext == ".puml" {
			content, err := os.ReadFile(absFilePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"type":    "uml",
				"content": string(content),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// For other files, return as text (could be extended for other types)
		content, err := os.ReadFile(absFilePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"type":    "text",
			"content": string(content),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// API to execute terminal commands via WebSocket streaming
	http.HandleFunc("/api/terminal/stream", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("WebSocket connection request received")

		// Upgrade HTTP connection to WebSocket
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all connections for now
			},
		}

		// Check if this is a WebSocket upgrade request
		if r.Header.Get("Upgrade") != "websocket" {
			fmt.Printf("Not a WebSocket upgrade request. Headers: %+v\n", r.Header)
			http.Error(w, "Expected WebSocket upgrade", http.StatusBadRequest)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Failed to upgrade to WebSocket: %v\n", err)
			return
		}
		defer ws.Close()
		fmt.Println("WebSocket connection established")

		// Get or create bash session
		session, err := initBashSession(absDir)
		if err != nil {
			fmt.Printf("Failed to create bash session: %v\n", err)
			ws.WriteJSON(map[string]string{"error": "Failed to create bash session: " + err.Error()})
			return
		}
		fmt.Println("Bash session obtained")

		// Create a context for this connection
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Handle incoming messages from WebSocket (user input)
		go func() {
			defer cancel()
			for {
				var msg map[string]string
				if err := ws.ReadJSON(&msg); err != nil {
					fmt.Printf("Error reading WebSocket message: %v\n", err)
					return
				}

				fmt.Printf("Received WebSocket message: %+v\n", msg)

				if input, ok := msg["input"]; ok {
					fmt.Printf("Sending input to bash: %q\n", input)
					if err := session.sendInput(input); err != nil {
						fmt.Printf("Error sending input to bash: %v\n", err)
						return
					}
				}
			}
		}()

		fmt.Println("Starting output loop")
		// Listen for output from the bash session
		for {
			select {
			case output := <-session.outputChannel:
				fmt.Printf("Bash output: %q\n", output)
				if err := ws.WriteJSON(map[string]string{"output": output}); err != nil {
					fmt.Printf("Error writing output to WebSocket: %v\n", err)
					return
				}
			case errOutput := <-session.errorChannel:
				fmt.Printf("Bash error: %q\n", errOutput)
				if err := ws.WriteJSON(map[string]string{"error": errOutput}); err != nil {
					fmt.Printf("Error writing error to WebSocket: %v\n", err)
					return
				}
			case <-ctx.Done():
				fmt.Println("WebSocket context cancelled")
				return
			case <-time.After(30 * time.Second):
				fmt.Println("Sending keepalive")
				// Send keepalive
				if err := ws.WriteJSON(map[string]bool{"keepalive": true}); err != nil {
					fmt.Printf("Error sending keepalive: %v\n", err)
					return
				}
			}
		}
	})

	// Remove the old /api/terminal/input endpoint since we're using WebSocket

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("Serving directory preview at http://localhost:%d\n", port)
	fmt.Printf("Directory: %s\n", absDir)

	go func() {
		time.Sleep(1 * time.Second)
		web.OpenBrowser(fmt.Sprintf("http://localhost:%d", port))
	}()

	return server.ListenAndServe()
}

func buildFileTree(rootPath string) (*FileNode, error) {
	stat, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		Name:  filepath.Base(rootPath),
		Path:  rootPath,
		IsDir: stat.IsDir(),
	}

	if stat.IsDir() {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			// Skip hidden files and directories
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			childPath := filepath.Join(rootPath, entry.Name())
			childNode, err := buildFileTree(childPath)
			if err != nil {
				// Skip files that can't be read
				continue
			}
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}
