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

	fileName, err := createDockerfile(dockerfile.Content)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	command := exec.Command("hadolint", fileName)
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

func createDockerfile(dockerfile string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Printf("Create tempfile error %s\n", err)
		return "", err
	}
	defer tmpfile.Close()

	log.Println(tmpfile.Name(), dockerfile)

	if _, err := tmpfile.WriteString(dockerfile); err != nil {
		log.Printf("Write tempfile error %s\n", err)
		return tmpfile.Name(), err
	}

	return tmpfile.Name(), nil
}

func parse(lintContent string) map[string][]string {
	linter := make(map[string][]string)

	if len(lintContent) == 0 ||
		strings.HasPrefix(lintContent, "hadolint") {
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
	http.Handle("/api/dockerfile", cors(lint))

	log.Printf("Starting server at port 8000.")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
