package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/StevenZack/openurl"
	"github.com/StevenZack/tools/netToolkit"
)

var (
	port = flag.String("p", ":8080", "port")
	dir  = flag.String("d", ".", "dir to serve")
)

func main() {
	flag.Parse()

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(*dir))))
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
