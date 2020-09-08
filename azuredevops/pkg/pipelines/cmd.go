package pipelines

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstspipelines "github.com/microsoft/azure-devops-go-api/azuredevops/pipelines"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yangzuo0621/azure-devops-cmd/azuredevops/pkg/vstspat"
)

const (
	flagVerbose      = "verbose"
	flagVerboseShort = "v"
	flagPatEnvKey    = "pat-env-key"
	vstsURL          = "https://dev.azure.com/%s"
)

var (
	organization string
	project      string
	pipelineID   int
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

func pipelineClientForCommandLine(cmd *cobra.Command) (PipelineClient, error) {
	logger := cmdLogger(cmd)

	patProvider, err := patEnvProvider(cmd)
	if err != nil {
		return nil, err
	}

	return newPipelineClient(logger, patProvider, organization, project)
}

// CreateCommand creates a cobra command instance of pipelines.
func CreateCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "pipeline",
		Short: "Manage an azure devops pipeline",
	}

	c.PersistentFlags().StringVar(&organization, "organization", "", "Organization the pipeline belongs to (required)")
	c.PersistentFlags().StringVar(&project, "project", "", "Project the pipeline belongs to (required)")
	c.MarkPersistentFlagRequired("organization")
	c.MarkPersistentFlagRequired("project")

	c.PersistentFlags().String(flagPatEnvKey, "VSTS_PAT", "env variable name for VSTS PAT (personal access token)")
	c.PersistentFlags().BoolP(flagVerbose, flagVerboseShort, false, "verbose output")

	c.AddCommand(createListPipelinesCommand())
	c.AddCommand(createGetPipelineCommand())
	c.AddCommand(createPipelineRunCommand())

	return c
}

func createListPipelinesCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "list",
		Short:        "list the specified project's pipelines",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			pipelineClient, err := pipelineClientForCommandLine(cmd)
			if err != nil {
				return err
			}

			pipelines, err := pipelineClient.ListPipelines(ctx)

			if err != nil {
				return err
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", " ")
			return encoder.Encode(pipelines)
		},
	}
	return c
}

func createGetPipelineCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "show [pipeline-id]",
		Short:        "Get detailed information of the specified pipeline",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineID, err := strconv.ParseInt(args[0], 10, 64)

			ctx := context.Background()

			pipelineClient, err := pipelineClientForCommandLine(cmd)
			if err != nil {
				return err
			}

			pipeline, err := pipelineClient.GetPipelineByID(ctx, int(pipelineID))
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", " ")
			return encoder.Encode(pipeline)
		},
	}
	return c
}

func createPipelineRunCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "run",
		Short: "Manage run install of azure devops pipeline",
	}

	c.PersistentFlags().IntVar(&pipelineID, "pipeline-id", 0, "pipeline id to retrieve")
	c.MarkPersistentFlagRequired("pipeline-id")

	c.AddCommand(createListPipelineRunCommand())
	c.AddCommand(createGetPipelineRunCommand())
	c.AddCommand(createTriggerPipelineRunCommand())
	return c
}

func createListPipelineRunCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "list",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			envKey, _ := cmd.Flags().GetString(flagPatEnvKey)
			personalAccessToken := os.Getenv(envKey)
			if personalAccessToken == "" {
				return fmt.Errorf("empty VSTS PAT from env %s", envKey)
			}

			ctx := context.Background()

			organizationURL := fmt.Sprintf(vstsURL, organization)

			connection := vsts.NewPatConnection(organizationURL, personalAccessToken)

			// Create a client to interact with the Pipelines area
			pipelineClient := vstspipelines.NewClient(ctx, connection)

			runs, err := pipelineClient.ListRuns(ctx, vstspipelines.ListRunsArgs{
				Project:    &project,
				PipelineId: &pipelineID,
			})

			if err != nil {
				return fmt.Errorf("Get pipeline runs failed for pipeline %d, error: %v", pipelineID, err)
			}

			if runs != nil {
				fmt.Printf("Count = %v\n", len(*runs))
			}

			return nil
		},
	}

	return c
}

func createGetPipelineRunCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "get [build-id]",
		Short:        "Get detailed information of the specified pipeline run",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceRunID, err := strconv.ParseInt(args[0], 10, 64)
			runID := int(sourceRunID)

			envKey, _ := cmd.Flags().GetString(flagPatEnvKey)
			personalAccessToken := os.Getenv(envKey)
			if personalAccessToken == "" {
				return fmt.Errorf("empty VSTS PAT from env %s", envKey)
			}

			ctx := context.Background()

			organizationURL := fmt.Sprintf(vstsURL, organization)

			connection := vsts.NewPatConnection(organizationURL, personalAccessToken)

			// Create a client to interact with the Pipelines area
			pipelineClient := vstspipelines.NewClient(ctx, connection)

			run, err := pipelineClient.GetRun(ctx, vstspipelines.GetRunArgs{
				Project:    &project,
				PipelineId: &pipelineID,
				RunId:      &runID,
			})

			if err != nil {
				return fmt.Errorf("Get pipeline run %d failed for pipeline %d, error: %v", runID, pipelineID, err)
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", " ")
			return encoder.Encode(run)
		},
	}

	return c
}

func createTriggerPipelineRunCommand() *cobra.Command {
	var (
		branch        string
		extraVarPairs []string
	)

	c := &cobra.Command{
		Use:          "trigger",
		Short:        "Trigger a build for a pipeline",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			envKey, _ := cmd.Flags().GetString(flagPatEnvKey)
			personalAccessToken := os.Getenv(envKey)
			if personalAccessToken == "" {
				return fmt.Errorf("empty VSTS PAT from env %s", envKey)
			}

			if branch == "" {
				branch = "master"
			}

			variables := map[string]vstspipelines.Variable{}
			for _, v := range extraVarPairs {
				parts := strings.SplitN(v, "=", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				variables[key] = vstspipelines.Variable{
					Value: &value,
				}
			}

			ctx := context.Background()

			organizationURL := fmt.Sprintf(vstsURL, organization)

			connection := vsts.NewPatConnection(organizationURL, personalAccessToken)

			// Create a client to interact with the Pipelines area
			pipelineClient := vstspipelines.NewClient(ctx, connection)

			responseValue, err := pipelineClient.RunPipeline(ctx, vstspipelines.RunPipelineArgs{
				Project:    &project,
				PipelineId: &pipelineID,
				RunParameters: &vstspipelines.RunPipelineParameters{
					Resources: &vstspipelines.RunResourcesParameters{
						Repositories: &map[string]vstspipelines.RepositoryResourceParameters{
							"self": {
								RefName: &branch,
							},
						},
					},
					Variables: &variables,
				},
			})

			if err != nil {
				return fmt.Errorf("Trigger pipeline run failed for pipeline %d, error: %v", pipelineID, err)
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", " ")
			return encoder.Encode(responseValue)
		},
	}

	c.Flags().StringVar(&branch, "branch", "", "The branch to trigger")
	c.Flags().StringSliceVar(&extraVarPairs, "var", []string{}, "extra variables to use")

	return c
}
