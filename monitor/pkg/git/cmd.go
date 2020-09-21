package git

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/vstspat"
)

const (
	flagVerbose      = "verbose"
	flagVerboseShort = "v"
	flagPatEnvKey    = "pat-env-key"
	vstsURL          = "https://dev.azure.com/%s"
)

var (
	organization  string
	project       string
	respositoryID string
	pipelineID    int
)

func patEnvProvider(cmd *cobra.Command) (vstspat.PATProvider, error) {
	envKey, _ := cmd.Flags().GetString(flagPatEnvKey)
	provider := vstspat.NewPATEnvBackend(envKey)
	k, err := provider.GetPAT(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get VSTS PAT from env %s failed: %w", envKey, err)
	}
	if k == "" {
		return nil, fmt.Errorf("empty VSTS PAT from env %s", envKey)
	}
	return provider, nil
}

func cmdLogger(cmd *cobra.Command) *logrus.Logger {
	logger := logrus.New()
	verbose, _ := cmd.Flags().GetBool(flagVerbose)
	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	return logger
}

func gitClientForCommandLine(cmd *cobra.Command) (GitClient, error) {
	logger := cmdLogger(cmd)

	patProvider, err := patEnvProvider(cmd)
	if err != nil {
		return nil, err
	}

	return newGitClient(logger, patProvider, organization, project, respositoryID)
}

// CreateCommand creates a cobra command instance of azure devops git.
func CreateCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "git",
		Short: "Manage an azure devops git respository",
	}

	c.PersistentFlags().StringVar(&organization, "organization", "", "Organization the pipeline belongs to (required)")
	c.PersistentFlags().StringVar(&project, "project", "", "Project the pipeline belongs to (required)")
	c.MarkPersistentFlagRequired("organization")
	c.MarkPersistentFlagRequired("project")

	c.PersistentFlags().String(flagPatEnvKey, "VSTS_PAT", "env variable name for VSTS PAT (personal access token)")
	c.PersistentFlags().BoolP(flagVerbose, flagVerboseShort, false, "verbose output")

	c.AddCommand(createGitPushCommand())
	c.AddCommand(createGetGitRepositoryCommand())
	c.AddCommand(createCloneGitRepositoryCommand())
	return c
}

func createGitPushCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "push",
		Short:        "create a new branch from commit",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, err := gitClientForCommandLine(cmd)
			if err != nil {
				return err
			}

			gitRef, err := client.PushNewGitBranch(ctx)

			if err != nil {
				return err
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", " ")
			return encoder.Encode(gitRef)
		},
	}

	c.Flags().StringVar(&respositoryID, "respository", "", "repository name or id (required)")
	c.MarkFlagRequired("respository")

	return c
}

func createGetGitRepositoryCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "get",
		Short:        "get a git repository",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, err := gitClientForCommandLine(cmd)
			if err != nil {
				return err
			}

			repo, err := client.GetGitRepository(ctx)

			if err != nil {
				return err
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", " ")
			return encoder.Encode(repo)
		},
	}

	c.Flags().StringVar(&respositoryID, "respository", "", "repository name or id (required)")
	c.MarkFlagRequired("respository")

	return c
}

func createCloneGitRepositoryCommand() *cobra.Command {
	var (
		repoName string
		repoPath string
	)
	c := &cobra.Command{
		Use:          "clone",
		Short:        "clone git repository",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := gitClientForCommandLine(cmd)
			if err != nil {
				return err
			}

			if repoPath == "" {
				repoPath = "./"
			}

			return client.CloneRepository(ctx, repoPath, repoName)
		},
	}

	c.Flags().StringVar(&repoName, "repo", "", "repository name (required)")
	c.Flags().StringVar(&repoPath, "path", "", "repository root path")
	c.MarkFlagRequired("repo")
	return c
}
