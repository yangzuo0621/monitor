package cicd

// Data encapsulates the status information about whole CI/CD process
type Data struct {
	E2EID         int    `json:"e2e_id"`
	CommitID      string `json:"commit_id"`
	BuildID       int    `json:"build_id"`
	BuildStatus   string `json:"build_status"`
	ReleaseID     int    `json:"release_id"`
	ReleaseStatus string `json:"release_status"`
}
