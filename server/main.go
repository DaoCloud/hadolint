package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	linter := parse(lint, fileName)

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

const specialLine = "\" (line "

func parse(lintContent, fileName string) map[string][]string {
	linter := make(map[string][]string)

	if len(lintContent) == 0 {
		return linter
	}

	lints := strings.Split(lintContent, fileName)
	for _, lint := range lints {
		// log.Println(lint)
		var numAndLinter []string
		if strings.HasPrefix(lint, specialLine) {
			numAndLinter = strings.SplitN(lint, ":\n", 2)
		} else {
			numAndLinter = strings.SplitN(lint, " ", 2)
		}
		if len(numAndLinter) != 2 {
			continue
		}

		lineNumber := getNumber(numAndLinter[0])
		lineLinter := strings.Trim(numAndLinter[1], "\n")
		linter[lineNumber] = append(linter[lineNumber], lineLinter)
	}

	return linter
}

func getNumber(fileNumber string) string {
	// log.Printf("line number %s\n", fileNumber)
	if strings.HasPrefix(fileNumber, specialLine) {
		fileNumer := strings.TrimPrefix(fileNumber, specialLine)
		i := ""
		for _, r := range fileNumer {
			if (r >= '0') && (r <= '9') {
				i += string(r)
			} else {
				break
			}
		}
		return i
	}

	fileAndNumer := strings.Split(fileNumber, ":")
	if len(fileAndNumer) == 1 {
		return "0"
	}

	return fileAndNumer[1]
}

func main() {
	http.Handle("/api/dockerfile", cors(lint))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Starting server at port %s.", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
