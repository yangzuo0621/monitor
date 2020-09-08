package pipelines

import (
	"context"

	vstsbuild "github.com/microsoft/azure-devops-go-api/azuredevops/build"
)

// PipelineClient interface for managing azure devops pipelines
type PipelineClient interface {
	// ListPipelines lists build pipelines.
	ListPipelines(ctx context.Context) ([]*vstsbuild.BuildDefinitionReference, error)
}
