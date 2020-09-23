package releases

import (
	"context"

	vstsrelease "github.com/microsoft/azure-devops-go-api/azuredevops/release"
)

// ReleaseClient interface for managing azure devops releases
type ReleaseClient interface {
	GetReleaseByID(ctx context.Context, releaseID int) (*vstsrelease.Release, error)
	ListReleases(ctx context.Context, releaseIDs []int) ([]*vstsrelease.Release, error)
	CreateRelease(ctx context.Context, definitionID int, alias string, buildID string, buildNumber string, description string) (*vstsrelease.Release, error)
}
