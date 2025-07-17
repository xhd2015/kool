package uml

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/xhd2015/kool/pkgs/web"
)

//go:embed index.html
var indexHtml string

func Serve(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	port, err := web.FindAvailablePort(8080, 100)
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(indexHtml))
	})
	http.HandleFunc("/api/content", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(content)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	fmt.Printf("Serving UML preview at http://localhost:%d\n", port)
	go func() {
		time.Sleep(1 * time.Second)
		web.OpenBrowser(fmt.Sprintf("http://localhost:%d", port))
	}()

	return server.ListenAndServe()
}
