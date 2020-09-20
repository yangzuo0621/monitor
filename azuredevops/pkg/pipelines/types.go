package pipelines

import (
	"context"

	vstsbuild "github.com/microsoft/azure-devops-go-api/azuredevops/build"
	vstspipelines "github.com/microsoft/azure-devops-go-api/azuredevops/pipelines"
)

// PipelineClient interface for managing azure devops pipelines
type PipelineClient interface {
	// ListPipelines lists build pipelines.
	ListPipelines(ctx context.Context) ([]*vstsbuild.BuildDefinitionReference, error)

	// GetPipelineByID gets a pipeline by id.
	GetPipelineByID(ctx context.Context, id int) (*vstsbuild.BuildDefinition, error)

	// ListPipelineBuilds lists builds of pipeline.
	ListPipelineBuilds(ctx context.Context, pipelineID int) ([]*vstsbuild.Build, error)

	// GetPipelineBuildByID gets a build of pipeline by id
	GetPipelineBuildByID(ctx context.Context, id int) (*vstsbuild.Build, error)

	// TriggerPipelineBuild creates a build intance of specified pipeline.
	TriggerPipelineBuild(ctx context.Context, pipelineID int, branch string, variables []string) (*vstspipelines.Run, error)

	// QueueBuild creates a build instance of specified pipeline.
	QueueBuild(ctx context.Context, pipelineID int, branch string, commitID string, variables map[string]string) (*vstsbuild.Build, error)

	GetArtifactsByBuildID(ctx context.Context, buildID int) (*[]vstsbuild.BuildArtifact, error)
}
