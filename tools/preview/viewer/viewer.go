package viewer

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/kool/pkgs/web"
)

//go:embed index.html
var indexHtml string

//go:embed index.css
var indexCss string

//go:embed index.js
var indexJs string

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

func Serve(dir string) error {
	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	port, err := web.FindAvailablePort(8080, 100)
	if err != nil {
		return err
	}

	htmlContent := strings.ReplaceAll(indexHtml, "__CSS__", indexCss)
	htmlContent = strings.ReplaceAll(htmlContent, "__SCRIPT__", indexJs)
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

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
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
