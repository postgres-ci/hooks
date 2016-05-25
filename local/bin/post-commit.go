package main

import (
	"github.com/postgres-ci/hooks/git"

	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
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

	response, err := client.Do(req)

	if err != nil {

		fmt.Printf("Error when sending a commit to Postgres-CI: %v\n", err)

		return nil
	}

	if response.StatusCode != http.StatusOK {

		fmt.Printf("Error when sending a commit to Postgres-CI: %v\n", response.Status)
	}

	return nil
}
