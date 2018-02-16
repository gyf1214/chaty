package controller

import (
	"flag"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const prefix = "/static/"

var path = flag.String("static", "./static", "static path")

func static(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	if !strings.HasPrefix(uri, prefix) || strings.Contains(uri, "..") {
		http.Error(w, "", 400)
		return
	}

	rel := strings.TrimPrefix(uri, prefix)
	path := filepath.Join(*path, rel)
	if file, err := os.Open(path); err == nil {
		ctype := mime.TypeByExtension(filepath.Ext(path))
		w.Header().Set("Content-Type", ctype)
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, "", 400)
		}
	} else {
		http.Error(w, "", 400)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, prefix+"html/index.html", 302)
}

func init() {
	http.HandleFunc(prefix, static)
	http.HandleFunc("/", index)
}
