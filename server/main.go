package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
)

type Dockerfile struct {
	Content string `dockerfile`
}

func lint(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}

	dockerfile := Dockerfile{}
	err := json.NewDecoder(req.Body).Decode(&dockerfile)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	b := new(bytes.Buffer)
	command := exec.Command("hadolint", dockerfile.Content)
	command.Stdout = b
	command.Stderr = b
	if err := command.Start(); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := command.Wait(); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	lint := string(b.Bytes())
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(lint)
}

func main() {
	http.HandleFunc("/api/dockerfile", lint)
	log.Fatal(http.ListenAndServe(":80", nil))
}
