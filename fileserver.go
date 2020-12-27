package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/StevenZack/openurl"
	"github.com/StevenZack/tools/fileToolkit"
	"github.com/StevenZack/tools/netToolkit"
)

var (
	port        = flag.String("p", ":8080", "port")
	dir         = flag.String("d", ".", "dir to serve")
	useTemplate = flag.Bool("t", false, "Use template")
)

type fileHandler struct {
	root string
}

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()

	if *useTemplate {
		http.Handle("/", &fileHandler{root: *dir})
		fmt.Println("using template engine")
	} else {
		http.Handle("/", http.FileServer(http.Dir(*dir)))
	}

	for _, ip := range netToolkit.GetIPs(false) {
		fmt.Println("listened on ", ip+*port)
	}
	openurl.Open("http://localhost" + *port)
	e := http.ListenAndServe(*port, nil)
	if e != nil {
		fmt.Println("listen error:", e)
		return
	}
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store")
	file := filepath.Base(r.RequestURI)
	if r.RequestURI == "/" {
		file = "index.html"
	}

	fmt.Println(r.RequestURI)

	// check
	if !strings.HasSuffix(file, ".html") {
		http.ServeFile(w, r, filepath.Join(f.root, r.RequestURI))
		return
	}

	fs, e := fileToolkit.Walk(f.root, func(path string) bool {
		return strings.HasSuffix(path, ".html")
	})
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	t, e := template.ParseFiles(fs...)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	e = t.ExecuteTemplate(w, file, nil)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

}
