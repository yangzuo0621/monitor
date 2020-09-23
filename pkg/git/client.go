package git

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstsgit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/monitor/pkg/vstspat"
)

const (
	vstsResourceURL = "https://dev.azure.com/%s"
	gitRepoFormat   = "https://dev.azure.com/%s/%s/_git/%s"
)

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

func (c *gitClient) CloneRepository(ctx context.Context, workdir string, repoName string) error {
	pat, err := c.patProvider.GetPAT(ctx)
	if err != nil {
		return fmt.Errorf("get PAT: %w", err)
	}

	auth := fmt.Sprintf(":%s", pat)
	authBase64Token := base64.StdEncoding.EncodeToString([]byte(auth))
	gitRepo := fmt.Sprintf(gitRepoFormat, c.organization, c.project, repoName)
	gitCmd := exec.Command("git", "-c", fmt.Sprintf(`http.extraHeader=Authorization: Basic %s`, authBase64Token), "clone", gitRepo)

	gitCmd.Dir = workdir
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("Clone repo %s: %v", repoName, err)
	}

	return nil
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

func gitCloneRepo(organization string, project string, repo string, pat string, dir string) {

	auth := fmt.Sprintf(":%s", pat)
	authBase64Token := base64.StdEncoding.EncodeToString([]byte(auth))
	gitRepo := fmt.Sprintf(gitRepoFormat, organization, project, repo)
	gitCmd := exec.Command("git", "-c", fmt.Sprintf(`http.extraHeader=Authorization: Basic %s`, authBase64Token), "clone", gitRepo)

	gitCmd.Dir = dir
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		fmt.Printf("err: %v", err)
	}
}

var _ GitClient = (*gitClient)(nil)
