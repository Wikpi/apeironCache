package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type upload struct {
	User       string `json:"user"`
	Type       string `json:"type"`
	Size       string `json:"size"`
	UploadBody []byte `json:"uploadBody"`
}

type serverModel struct {
	mux *http.ServeMux

	users   map[string]string
	uploads map[string]upload
}

func newServer() *serverModel {
	model := &serverModel{}

	model.users = make(map[string]string)
	model.uploads = make(map[string]upload)

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

func (s *serverModel) handleUpload(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not read request body")
		return
	}
	var newUpload upload

	if err := json.Unmarshal(body, &newUpload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not parse request")
		return
	}

	code, err := generateCode(6)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not generate unique code")
		return
	}
	s.uploads[code] = newUpload

	data, err := json.Marshal(code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not parse client unique code to json")
		return
	}
	w.WriteHeader(http.StatusAccepted)

	w.Write(data)

	log.Println("Pasted at: ", code, ", with: ", newUpload)
}

func (s *serverModel) handleGet(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not read request body")
	}
	var code string

	if err := json.Unmarshal(body, &code); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not parse request")
	}

	getRequest := s.uploads[code]

	data, err := json.Marshal(getRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Could not parse client get request to json")
	}
	w.WriteHeader(http.StatusAccepted)

	w.Write(data)

	log.Println("Sent client get request via " + code + " code")
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
	server.mux.HandleFunc("/paste", server.handleUpload)
	server.mux.HandleFunc("/get", server.handleGet)

	if err := http.ListenAndServe("localhost:8080", server.mux); err != nil {
		log.Fatal("Could not start server")
	}
}
