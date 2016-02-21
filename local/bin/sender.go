package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/postgres-ci/hooks/git"
)

type Commit struct {
	Ref string `json:"ref"`
	git.Commit
}

func main() {

	ref, err := git.GetCurrentRef(os.Getenv("PWD"))

	if err != nil {

		return
	}

	if commit, err := git.GetLastCommit(os.Getenv("PWD")); err == nil {

		send(Commit{
			Ref:    ref,
			Commit: commit,
		})
	}
}

var client http.Client

func send(commit Commit) error {

	//data, _ := json.Marshal(push)

	data, _ := json.MarshalIndent(commit, "  ", "  ")

	fmt.Println(string(data))

	_url := url.URL{
		Scheme: "http",
		Host:   os.Getenv("API_HOST"),
		Path:   "/web-hooks/native/",
	}

	req, _ := http.NewRequest("POST", _url.String(), bytes.NewReader(data))

	req.Header.Set("X-Event", "commit")
	req.Header.Set("X-Repository", os.Getenv("REPOSITORY"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := client.Do(req); err != nil {

		return nil
	}

	return nil
}
