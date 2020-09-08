package pipelines

import (
	"context"
	"fmt"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstsbuild "github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/azure-devops-cmd/azuredevops/pkg/vstspat"
)

const vstsResourceURL = "https://dev.azure.com/%s"

type pipelineClient struct {
	patProvider  vstspat.PATProvider
	organization string
	project      string

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

func newPipelineClient(rootLogger logrus.FieldLogger, patProvider vstspat.PATProvider, org string, project string) (PipelineClient, error) {
	logger := rootLogger.WithFields(logrus.Fields{
		"organization": org,
		"project":      project,
	})

	return &pipelineClient{
		patProvider:  patProvider,
		organization: org,
		project:      project,
		logger:       logger,
	}, nil
}

var _ PipelineClient = (*pipelineClient)(nil)
