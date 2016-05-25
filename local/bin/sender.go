package main

import (
	"github.com/postgres-ci/hooks/git"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
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

var client = http.Client{
	Timeout: time.Second * 2,
}

func send(commit Commit) error {

	data, _ := json.Marshal(commit)

	_url := url.URL{
		Scheme: os.Getenv("SCHEME"),
		Host:   os.Getenv("HOST"),
		Path:   "/webhooks/native/",
	}

	req, _ := http.NewRequest("POST", _url.String(), bytes.NewReader(data))

	req.Header.Set("X-Event", "commit")
	req.Header.Set("X-Token", os.Getenv("TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	if _, err := client.Do(req); err != nil {

		fmt.Printf("Error when sending a commit to Postgres-CI (err: %v)\n", err)

		return nil
	}

	fmt.Printf("\nCommit %s was sent to Postgres-CI server (%s://%s)\n\n", commit.ID, _url.Scheme, _url.Host)

	return nil
}
