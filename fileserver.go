package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/StevenZack/mux"
	"github.com/StevenZack/openurl"
	"github.com/StevenZack/tools/fileToolkit"
	"github.com/StevenZack/tools/netToolkit"
)

var (
	port        = flag.String("p", "8080", "port")
	dir         = flag.String("d", ".", "dir to serve")
	useTemplate = flag.Bool("t", false, "Use template")
	index       = flag.Bool("i", false, "Route 404 to index.html")
	s           *mux.Server
	v           = flag.Bool("v", false, "version")
)

const (
	currentVersion = "v1.0.0"
)

type fileHandler struct {
	root string
}

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()
	if *v {
		fmt.Println(currentVersion)
		return
	}
	s = mux.NewServer(":" + *port)
	if *useTemplate {
		s.Handle("/", &fileHandler{root: *dir})
		fmt.Println("using template engine")
	} else {
		s.HandleMultiReqs("/", fs)
	}

	for _, ip := range netToolkit.GetIPs(false) {
		fmt.Println("listened on ", ip+":"+*port)
	}
	openurl.Open("http://localhost:" + *port)
	e := s.ListenAndServe()
	if e != nil {
		fmt.Println("listen error:", e)
		return
	}
}

func fs(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if p == "/" {
		p = "/index.html"
	}
	path := filepath.Join(*dir, p)
	info, e := os.Stat(path)
	if e != nil {
		if os.IsNotExist(e) {
			if *index {
				http.ServeFile(w, r, filepath.Join(*dir, "index.html"))
				return
			}
			s.NotFound(w, r)
			return
		}
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		path = filepath.Join(path, "index.html")
	}
	// w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
	http.ServeFile(w, r, path)
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
