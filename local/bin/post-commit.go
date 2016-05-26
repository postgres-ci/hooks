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

		err := send(Commit{
			Ref:    ref,
			Commit: commit,
		})

		if err != nil {

			fmt.Println(err)
		}
	}
}

var client = http.Client{

	Timeout: time.Second * 2,
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

		return fmt.Errorf("Error when sending a commit to Postgres-CI: %v", err)
	}

	if response.StatusCode != http.StatusOK {

		if response.StatusCode == http.StatusBadRequest {

			var message struct {
				Success bool   `json:"success"`
				Code    int    `json:"code"`
				Error   string `json:"error"`
			}

			if err := json.NewDecoder(response.Body).Decode(&message); err == nil {

				return fmt.Errorf("Error when sending a commit to Postgres-CI: %s", message.Error)
			}
		}

		return fmt.Errorf("Error when sending a commit to Postgres-CI: %s", response.Status)
	}

	fmt.Printf("\nCommit %s was sent to Postgres-CI server (%s://%s)\n\n", commit.ID, _url.Scheme, _url.Host)

	return nil
}
