package pipelines

import (
	"context"
	"fmt"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
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
