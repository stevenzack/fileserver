package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stevenzack/openurl"
)

var (
	port        = flag.String("p", "8080", "port")
	dir         = flag.String("d", ".", "dir to serve")
	useTemplate = flag.Bool("t", false, "Use template")
	index       = flag.Bool("i", false, "Route 404 to index.html")
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/") {
			p += "index.html"
		}
		p = strings.TrimPrefix(p, "/")
		dst := p
		if *dir != "" {
			filepath.Join(*dir, dst)
		}

		if _, e := os.Stat(dst); os.IsNotExist(e) {
			if *index {
				http.ServeFile(w, r, filepath.Join(*dir, "index.html"))
				return
			}

			http.NotFound(w, r)
			return
		}

		if *useTemplate && filepath.Ext(p) == ".html" {

			var tp *template.Template
			e := filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
				ext := filepath.Ext(path)
				if ext != ".html" {
					return nil
				}
				rel, e := filepath.Rel(*dir, path)
				if e != nil {
					log.Println(e)
					return e
				}
				b, e := os.ReadFile(path)
				if e != nil {
					log.Println(e)
					return e
				}

				if tp == nil {
					tp = template.New(rel)
				} else {
					tp = tp.New(rel)
				}
				tp, e = tp.Parse(string(b))
				if e != nil {
					log.Println(e)
					return e
				}

				return nil
			})
			if e != nil {
				http.Error(w, fmt.Sprintf(`{"code":3,"desc":"%s"}`, e.Error()), http.StatusBadRequest)
				return
			}
			if tp != nil {
				w.Header().Set("Content-Type", "text/html")
				e = tp.ExecuteTemplate(w, p, nil)
				if e != nil {
					w.Header().Set("Content-Type", "application/json")
					http.Error(w, fmt.Sprintf(`{"code":3,"desc":"%s"}`, e.Error()), http.StatusBadRequest)
					return
				}
				return
			}
		}

		http.ServeFile(w, r, dst)
	})

	for _, ip := range GetIPs(false) {
		println("listened on ", ip+":"+*port)
	}
	openurl.Open("http://localhost:" + *port)
	e := http.ListenAndServe(":"+*port, nil)
	if e != nil {
		log.Panic(e)
		return
	}

}

func GetIPs(ipv6 bool) []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	strs := []string{}
	maxAddr := ""
	maxValue := 0
	maxIndex := 0
	ipv6s := []string{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip := v.IP
				if strings.HasSuffix(ip.String(), "::1") || ip.String() == "127.0.0.1" {
					continue
				}
				if strings.Contains(ip.String(), ":") {
					if ipv6 {
						ipv6s = append(ipv6s, "["+ip.String()+"]")
					}
					continue
				}
				strs = append(strs, ip.String())
				value, e := strconv.Atoi(SubBefore(ip.String(), ".", "0"))
				if e != nil {
					continue
				}
				if value > maxValue {
					maxValue = value
					maxAddr = ip.String()
					maxIndex = len(strs) - 1
				}
			case *net.IPAddr:
				// ip := v.IP
				// strs = append(strs, ip.String())
			}
		}
	}
	if len(strs) == 0 {
		return nil
	}
	strs = append([]string{maxAddr}, append(strs[:maxIndex], strs[maxIndex+1:]...)...)
	return append(strs, ipv6s...)
}

func SubBefore(s, sep, def string) string {
	if len(sep) > len(s) {
		return def
	}
	for i := range s {
		if s[i:i+len(sep)] == sep {
			return s[:i]
		}
	}
	return def
}
