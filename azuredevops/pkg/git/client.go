package git

import (
	"context"
	"fmt"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstsgit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/azure-devops-cmd/azuredevops/pkg/vstspat"
)

const vstsResourceURL = "https://dev.azure.com/%s"

type gitClient struct {
	patProvider  vstspat.PATProvider
	organization string
	project      string
	repositoryID string

	logger logrus.FieldLogger
}

func (c *gitClient) patTokenConn(ctx context.Context) (*vsts.Connection, error) {
	pat, err := c.patProvider.GetPAT(ctx)
	if err != nil {
		return nil, fmt.Errorf("get PAT: %w", err)
	}
	organizationURL := fmt.Sprintf(vstsResourceURL, c.organization)
	conn := vsts.NewPatConnection(organizationURL, pat)

	return conn, nil
}

func (c *gitClient) buildClient(ctx context.Context) (vstsgit.Client, error) {
	connection, err := c.patTokenConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire pat connection: %w", err)
	}

	client, err := vstsgit.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("new policy client: %w", err)
	}

	return client, nil
}

func (c *gitClient) PushNewGitBranch(ctx context.Context) (*vstsgit.GitRef, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "pushNewGitBranch",
	})

	client, err := c.buildClient(ctx)
	if err != nil {
		logger.WithError(err).Error()
		return nil, err
	}

	newObjectID := "0000000000000000000000000000000000000000"
	oldObjectID := "23efe240e33f426fb3364d71bafcf6ff0a440914"
	branch := "refs/heads/zuya/test-branch"
	master := "/refs/heads/zuya/e2e-cluster-verify-handler"
	isLocked := false
	resp, err := client.UpdateRef(ctx, vstsgit.UpdateRefArgs{
		NewRefInfo: &vstsgit.GitRefUpdate{
			OldObjectId: &oldObjectID,
			NewObjectId: &newObjectID,
			Name:        &branch,
			IsLocked:    &isLocked,
		},
		RepositoryId: &c.repositoryID,
		Filter:       &master,
		Project:      &c.project,
	})

	if err != nil {
		err = fmt.Errorf("push get ref failed: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}
	return resp, nil
}

func (c *gitClient) GetGitRepository(ctx context.Context) (*vstsgit.GitRepository, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "getGitRepository",
	})

	client, err := c.buildClient(ctx)
	if err != nil {
		logger.WithError(err).Error()

		return nil, err
	}

	repo, err := client.GetRepository(ctx, vstsgit.GetRepositoryArgs{
		RepositoryId: &c.repositoryID,
		Project:      &c.project,
	})

	if err != nil {
		return nil, err
	}

	return repo, nil
}

func newGitClient(rootLogger logrus.FieldLogger, patProvider vstspat.PATProvider, org string, project string, repositoryID string) (GitClient, error) {
	logger := rootLogger.WithFields(logrus.Fields{
		"organization": org,
		"repositoryID": repositoryID,
	})

	return &gitClient{
		patProvider:  patProvider,
		organization: org,
		project:      project,
		repositoryID: repositoryID,
		logger:       logger,
	}, nil
}

var _ GitClient = (*gitClient)(nil)
