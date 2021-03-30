package main

import (
	"embed"
	_ "embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed html/*.html
var htmlFs embed.FS

//go:embed static/*
var static embed.FS

func main() {
	html, err := fs.Sub(htmlFs, "html")
	if err != nil {
		log.Fatal(err)
	}
	t, err := template.ParseFS(html, "*")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/static/", http.FileServer(http.FS(static)))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		pp := strings.Split(req.URL.Path, "/")
		path := pp[len(pp)-1]
		log.Println("Serving request for path", path)
		w.Header().Add("Content-Type", "text/html")
		// respond with the output of template execution
		if path == "" {
			path = "index"
		}
		if err := t.ExecuteTemplate(w, path+".html", nil); err != nil {
			w.WriteHeader(http.StatusNotFound)
		}

	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
