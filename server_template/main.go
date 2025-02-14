package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/xhd2015/kool/server_template/route"
)

func main() {
	// serve favicon.ico
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/icon.svg")
	})

	// serve frontend/index.js
	indexHandler, err := route.NewFileHandler("../frontend/build/index.js")
	if err != nil {
		log.Fatalf("Failed to initialize index handler: %v", err)
	}
	http.HandleFunc("/index.js", indexHandler.ServeHTTP)

	// serve frontend/template.html
	templateHandler, err := route.NewFileHandler("../frontend/template.html")
	if err != nil {
		log.Fatalf("Failed to initialize template handler: %v", err)
	}
	helloHandler := templateHandler.Clone()
	http.HandleFunc("/hello.html", helloHandler.AddProcess(func(content []byte) ([]byte, error) {
		content = bytes.ReplaceAll(content, []byte("__TITLE__"), []byte("Hello"))
		content = bytes.ReplaceAll(content, []byte("__RENDER__"), []byte("renderRoute"))
		content = bytes.ReplaceAll(content, []byte("__COMPONENT__"), []byte("AppRoutes"))
		content = bytes.ReplaceAll(content, []byte("__INDEX_PATH__"), []byte(""))
		return content, nil
	}).ServeHTTP)

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		// refresh helleHandler
		err := helloHandler.Refresh()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = indexHandler.Refresh()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok\n"))
	})

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
