package main

import (
	"crypto/rand"
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
	pastes map[string]paste
}

func newServer() *serverModel {
	model := &serverModel{}

	model.users = make(map[string]string)
	model.pastes = make(map[string]paste)

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
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not read request body")
	}
	var pastedMessage paste

	if err := json.Unmarshal(body, &pastedMessage); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not parse request")
	}

	code, err := generateCode(6)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not generate unique code")
	}
	w.WriteHeader(http.StatusAccepted)

	data, err := json.Marshal(code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not parse client unique code to json")
	}

	w.Write(data)

	log.Println("Pasted at: ", code, ", with: ", pastedMessage)
}

func generateCode(length int) (string, error) {
	const codeChars = "1234567890"

	buffer := make([]byte, length)

	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	for index := range buffer {
		buffer[index] = codeChars[int(buffer[index])%len(codeChars)]
	}

	return string(buffer), nil
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
