package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type Dockerfile struct {
	Content string `json:"dockerfile"`
}

func lint(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dockerfile := Dockerfile{}
	err := json.NewDecoder(req.Body).Decode(&dockerfile)
	if err != nil {
		log.Printf("Decode body error %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	tmpfile, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Printf("Create tempfile error %s\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer tmpfile.Close()

	log.Println(tmpfile.Name(), dockerfile.Content)

	if _, err := tmpfile.WriteString(dockerfile.Content); err != nil {
		log.Printf("Write tempfile error %s\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	command := exec.Command("hadolint", tmpfile.Name())
	var out bytes.Buffer
	command.Stdout = &out
	command.Stderr = &out

	command.Run()

	lint := string(out.Bytes())
	log.Println(lint)

	linter := parse(lint)

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(linter)
}

func parse(lintContent string) map[string][]string {
	linter := make(map[string][]string)

	if len(lintContent) == 0 {
		return linter
	}

	lints := strings.Split(lintContent, "\n")
	for _, lint := range lints {
		numAndlinter := strings.SplitN(lint, " ", 2)
		if len(numAndlinter) != 2 {
			continue
		}

		lineNumber := getNumber(numAndlinter[0])
		lineLinter := numAndlinter[1]
		linter[lineNumber] = append(linter[lineNumber], lineLinter)
	}

	return linter
}

func getNumber(fileNumber string) string {
	fileAndNumer := strings.Split(fileNumber, ":")
	if len(fileAndNumer) == 1 {
		return "0"
	}

	return fileAndNumer[1]
}

func main() {
	http.HandleFunc("/api/dockerfile", lint)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
