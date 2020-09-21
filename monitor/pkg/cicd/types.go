package cicd

// Data encapsulates the status information about whole CI/CD process
type Data struct {
	E2EID         int    `json:"e2e_id"`
	CommitID      string `json:"commit_id"`
	BuildID       int    `json:"build_id"`
	BuildStatus   string `json:"build_status"`
	BuildResult   string `json:"build_result"`
	ReleaseID     int    `json:"release_id"`
	ReleaseDate   string `json:"release_date"`
	ReleaseStatus string `json:"release_status"`
}
