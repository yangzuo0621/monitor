package cicd

// Data encapsulates the status information about whole CI/CD process
type Data struct {
	MasterValidation *MasterValidation `json:"e2e_master_validation,omitempty"`
	AKSBuild         *AKSBuild         `json:"ev2_aks_build,omitempty"`
	AKSRelease       *AKSRelease       `json:"ev2_aks_release,omitempty"`
	State            DataState         `json:"state"`
}

// MasterValidation encapsulates the information about `E2Ev2 AKS RP Master Validation`
type MasterValidation struct {
	ID       int    `json:"id"`
	CommitID string `json:"commit_id"`
	Branch   string `json:"branch"`
}

// AKSBuild encapsulates the information about `[EV2] AKS Build` runs
type AKSBuild struct {
	ID          int     `json:"id"`
	BuildStatus *string `json:"status,omitempty"`
	BuildResult *string `json:"result,omitempty"`
	Count       int     `json:"count"`
}

// AKSRelease encapsulates the information about `AKS Release` runs
type AKSRelease struct {
	ID int `json:"id"`
}

type DataState string

type dataStateValuesType struct {
	None              DataState
	NotStart          DataState
	BuildInProgress   DataState
	BuildFailed       DataState
	BuildSucceeded    DataState
	ReleaseInProgress DataState
	ReleaseFailed     DataState
	ReleaseSucceeded  DataState
}

var DataStateValues = dataStateValuesType{
	None:              "none",
	NotStart:          "notStart",
	BuildInProgress:   "buildInProgress",
	BuildFailed:       "buildFailed",
	BuildSucceeded:    "buildSucceeded",
	ReleaseInProgress: "releaseInProgress",
	ReleaseFailed:     "releaseFailed",
	ReleaseSucceeded:  "releaseSucceeded",
}
