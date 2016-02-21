package git

import "time"

type Push struct {
	Ref     string   `json:"ref"`
	Old     string   `json:"old"`
	New     string   `json:"new"`
	Commits []Commit `json:"commits"`
}

type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Commit struct {
	ID          string    `json:"id"`
	Author      Committer `json:"author"`
	Committer   Committer `json:"committer"`
	Message     string    `json:"message"`
	CommittedAt time.Time `json:"committed_at"`
}
