package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type paste struct {
	Name      string `json:"name"`
	PasteBody string `json:"pasteBody"`
}

type serverModel struct {
	mux *http.ServeMux

	users  map[string]string
	pastes []paste
}

func newServer() *serverModel {
	model := &serverModel{}

	model.users = make(map[string]string)

	model.mux = http.NewServeMux()

	return model
}

func (s *serverModel) registerUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		log.Println("Could not read request body")
	}
	var name string

	if err := json.Unmarshal(body, &name); err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		log.Println("Could not parse request")
	}

	if _, ok := s.users[name]; ok {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	s.users[name] = ""

	log.Println("New user ", name)
}

func (s *serverModel) handlePaste(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		log.Println("Could not read request body")
	}
	var pastedMessage paste

	if err := json.Unmarshal(body, &pastedMessage); err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		log.Println("Could not parse request")
	}

	s.pastes = append(s.pastes, pastedMessage)
	w.WriteHeader(http.StatusAccepted)

	log.Println(pastedMessage)
}

func main() {
	server := newServer()

	fmt.Println("Running on localhost:8080")

	server.mux.HandleFunc("/register", server.registerUser)
	server.mux.HandleFunc("/paste", server.handlePaste)

	if err := http.ListenAndServe("localhost:8080", server.mux); err != nil {
		log.Fatal("Could not start server")
	}
}
