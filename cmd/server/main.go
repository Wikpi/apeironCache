package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// type serverModel struct {
// }

// func newServer() *serverModel {
// 	model := &serverModel{}

// 	return model
// }

func serveRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Got root request\n")
	io.WriteString(w, "Server is working\n")
}

func main() {
	//server := newServer()

	http.HandleFunc("/", serveRoot)

	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal("Could not start server")
	}
}
