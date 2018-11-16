package main

import (
	"fmt"
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
	e := http.ListenAndServe(port, nil)
	if e != nil {
		fmt.Println("listen error:", e)
		return
	}
}
