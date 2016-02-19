package git

type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type Commit struct {
	ID        string    `json:"id"`
	Author    Committer `json:"author"`
	Committer Committer `json:"committer"`
	Message   string    `json:"message"`
}

type Push struct {
	Repository string   `json:"repository"`
	Ref        string   `json:"ref"`
	Before     string   `json:"before"`
	After      string   `json:"after"`
	Commits    []Commit `json:"commits"`
}
