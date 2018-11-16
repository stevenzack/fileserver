package main

import (
	"fmt"
	"github.com/StevenZack/tools/netToolkit"
	"net/http"
	"os"
)

func main() {
	dir := "."
	port := ":8080"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if len(os.Args) > 2 {
		port = os.Args[2]
	}
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(dir))))
	for _, ip := range netToolkit.GetIPs() {
		fmt.Println("listened on ", ip+port)
	}
	e := http.ListenAndServe(port, nil)
	if e != nil {
		fmt.Println("listen error:", e)
		return
	}
}
