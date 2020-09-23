package releases

import (
	"context"
	"fmt"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstsrelease "github.com/microsoft/azure-devops-go-api/azuredevops/release"
	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/monitor/pkg/vstspat"
)

const vstsResourceURL = "https://dev.azure.com/%s"

type releaseClient struct {
	patProvider  vstspat.PATProvider
	organization string
	project      string

	logger logrus.FieldLogger
}

func (c *releaseClient) patConnection(ctx context.Context) (*vsts.Connection, error) {
	pat, err := c.patProvider.GetPAT(ctx)
	if err != nil {
		return nil, fmt.Errorf("get PAT: %w", err)
	}

	organizationURL := fmt.Sprintf(vstsResourceURL, c.organization)
	conn := vsts.NewPatConnection(organizationURL, pat)
	return conn, nil
}

func (c *releaseClient) buildClient(ctx context.Context) (vstsrelease.Client, error) {
	connection, err := c.patConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire pat connection: %w", err)
	}

	client, err := vstsrelease.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("new policy client: %w", err)
	}

	return client, nil
}

func (c *releaseClient) GetReleaseByID(ctx context.Context, releaseID int) (*vstsrelease.Release, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "ListReleases",
	})

	client, err := c.buildClient(ctx)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	release, err := client.GetRelease(ctx, vstsrelease.GetReleaseArgs{
		Project:   &c.project,
		ReleaseId: &releaseID,
	})
	if err != nil {
		err := fmt.Errorf("get release %d: %w", releaseID, err)
		logger.WithError(err).Error()
		return nil, err
	}

	return release, nil

}

// func (c *releaseClient) ListReleases(ctx context.Context) ([]*vstsrelease.Release, error) {
// 	logger := c.logger.WithFields(logrus.Fields{
// 		"action": "ListReleases",
// 	})

// 	client, err := c.buildClient(ctx)
// 	if err != nil {
// 		logger.WithError(err).Error()
// 		return nil, err
// 	}

// 	resp, err := client.GetReleases(ctx, vstsrelease.GetReleasesArgs{})
// 	if err != nil {
// 		err = fmt.Errorf("get Releases failed: %w", err)
// 		logger.WithError(err).Error()
// 		return nil, err
// 	}

// 	var result []*vstsrelease.Release
// 	for _, v := range resp.Value {
// 		value := v
// 		result = append(result, &value)
// 	}

// 	return result, nil
// }

func BuildReleaseClient(rootLogger logrus.FieldLogger, patProvider vstspat.PATProvider, org string, project string) (ReleaseClient, error) {
	logger := rootLogger.WithFields(logrus.Fields{
		"organization": org,
		"project":      project,
	})

	return &releaseClient{
		patProvider:  patProvider,
		organization: org,
		project:      project,
		logger:       logger,
	}, nil
}

var _ ReleaseClient = (*releaseClient)(nil)
