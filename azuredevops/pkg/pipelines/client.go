package pipelines

import (
	"context"
	"fmt"
	"strings"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstsbuild "github.com/microsoft/azure-devops-go-api/azuredevops/build"
	vstspipelines "github.com/microsoft/azure-devops-go-api/azuredevops/pipelines"
	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/azure-devops-cmd/azuredevops/pkg/vstspat"
)

const vstsResourceURL = "https://dev.azure.com/%s"

type pipelineClient struct {
	patProvider  vstspat.PATProvider
	organization string
	project      string
	pipelineID   int

	logger logrus.FieldLogger
}

func (c *pipelineClient) patTokenConn(ctx context.Context) (*vsts.Connection, error) {
	pat, err := c.patProvider.GetPAT(ctx)
	if err != nil {
		return nil, fmt.Errorf("get PAT: %w", err)
	}
	organizationURL := fmt.Sprintf(vstsResourceURL, c.organization)
	conn := vsts.NewPatConnection(organizationURL, pat)

	return conn, nil
}

func (c *pipelineClient) buildClient(ctx context.Context) (vstsbuild.Client, error) {
	connection, err := c.patTokenConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire pat connection: %w", err)
	}

	client, err := vstsbuild.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("new policy client: %w", err)
	}

	return client, nil
}

func (c *pipelineClient) pipelineClient(ctx context.Context) (vstspipelines.Client, error) {
	connection, err := c.patTokenConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire pat connection: %w", err)
	}

	client := vstspipelines.NewClient(ctx, connection)
	return client, nil
}

func (c *pipelineClient) ListPipelines(ctx context.Context) ([]*vstsbuild.BuildDefinitionReference, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "listPipelines",
	})

	buildClient, err := c.buildClient(ctx)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	resp, err := buildClient.GetDefinitions(ctx, vstsbuild.GetDefinitionsArgs{
		Project: &c.project,
	})

	if err != nil {
		err = fmt.Errorf("get definitions failed: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	var result []*vstsbuild.BuildDefinitionReference
	for _, v := range resp.Value {
		value := v
		result = append(result, &value)
	}

	return result, nil
}

func (c *pipelineClient) GetPipelineByID(ctx context.Context, id int) (*vstsbuild.BuildDefinition, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action":      "getPipelineByID",
		"pipeline.id": id,
	})

	buildClient, err := c.buildClient(ctx)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	pipeline, err := buildClient.GetDefinition(ctx, vstsbuild.GetDefinitionArgs{
		Project:      &c.project,
		DefinitionId: &id,
	})
	if err != nil {
		err := fmt.Errorf("get build definition %d: %w", id, err)
		logger.WithError(err).Error()
		return nil, err
	}

	return pipeline, nil
}

func (c *pipelineClient) ListPipelineBuilds(ctx context.Context) ([]*vstsbuild.Build, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action":      "listPipelineBuilds",
		"pipeline.id": c.pipelineID,
	})

	buildClient, err := c.buildClient(ctx)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	resp, err := buildClient.GetBuilds(ctx, vstsbuild.GetBuildsArgs{
		Project:     &c.project,
		Definitions: &[]int{c.pipelineID},
	})

	if err != nil {
		err = fmt.Errorf("get builds failed: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	var result []*vstsbuild.Build
	for _, v := range resp.Value {
		value := v
		result = append(result, &value)
	}

	return result, nil
}

func (c *pipelineClient) GetPipelineBuildByID(ctx context.Context, id int) (*vstsbuild.Build, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action":      "listPipelineBuilds",
		"pipeline.id": c.pipelineID,
		"build.id":    id,
	})

	buildClient, err := c.buildClient(ctx)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	build, err := buildClient.GetBuild(ctx, vstsbuild.GetBuildArgs{
		Project: &c.project,
		BuildId: &id,
	})

	if err != nil {
		err = fmt.Errorf("get build %d failed: %w", id, err)
		logger.WithError(err).Error()
		return nil, err
	}

	return build, nil
}

func (c *pipelineClient) TriggerPipelineBuild(ctx context.Context, branch string, variables []string) (*vstspipelines.Run, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action":      "listPipelineBuilds",
		"pipeline.id": c.pipelineID,
	})

	pipelineClient, err := c.pipelineClient(ctx)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	if branch == "" {
		branch = "master"
	}

	vars := map[string]vstspipelines.Variable{}
	for _, v := range variables {
		parts := strings.SplitN(v, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		vars[key] = vstspipelines.Variable{
			Value: &value,
		}
	}

	resp, err := pipelineClient.RunPipeline(ctx, vstspipelines.RunPipelineArgs{
		Project:    &c.project,
		PipelineId: &c.pipelineID,
		RunParameters: &vstspipelines.RunPipelineParameters{
			Resources: &vstspipelines.RunResourcesParameters{
				Repositories: &map[string]vstspipelines.RepositoryResourceParameters{
					"self": {
						RefName: &branch,
					},
				},
			},
			Variables: &vars,
		},
	})

	if err != nil {
		err = fmt.Errorf("trigger pipeline %d failed: %w", c.pipelineID, err)
		logger.WithError(err).Error()
		return nil, err
	}

	return resp, nil
}

func newPipelineClient(rootLogger logrus.FieldLogger, patProvider vstspat.PATProvider, org string, project string, pipelineID int) (PipelineClient, error) {
	logger := rootLogger.WithFields(logrus.Fields{
		"organization": org,
		"project":      project,
	})

	return &pipelineClient{
		patProvider:  patProvider,
		organization: org,
		project:      project,
		pipelineID:   pipelineID,
		logger:       logger,
	}, nil
}

var _ PipelineClient = (*pipelineClient)(nil)
