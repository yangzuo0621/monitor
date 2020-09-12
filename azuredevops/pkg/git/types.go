package git

import (
	"context"

	vstsgit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

// GitClient interface for managing azure devops git
type GitClient interface {
	// PushNewGitBranch create a git branch
	PushNewGitBranch(ctx context.Context) (*vstsgit.GitRef, error)

	// GetGitRepository get git repository
	GetGitRepository(ctx context.Context) (*vstsgit.GitRepository, error)

	// CloneRepository clones a repository
	CloneRepository(ctx context.Context, workdir string, repoName string) error
}
