package releases

import (
	"context"

	vstsrelease "github.com/microsoft/azure-devops-go-api/azuredevops/release"
)

type ReleaseClient interface {
	GetReleaseByID(ctx context.Context, releaseID int) (*vstsrelease.Release, error)
}
