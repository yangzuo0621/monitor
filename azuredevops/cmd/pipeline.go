package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstspipelines "github.com/microsoft/azure-devops-go-api/azuredevops/pipelines"
	"github.com/spf13/cobra"
)

var (
	organization        string
	project             string
	pipelineID          int
	runID               int
	personalAccessToken string
)

func init() {
	pipelinesCmd.PersistentFlags().StringVarP(&organization, "organization", "o", "", "Organization the pipeline belongs to (required)")
	pipelinesCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project the pipeline belongs to (required)")
	pipelinesCmd.PersistentFlags().StringVarP(&personalAccessToken, "token", "t", "", "Personal access token to azure devops")
	pipelinesCmd.MarkPersistentFlagRequired("organization")
	pipelinesCmd.MarkPersistentFlagRequired("project")

	listPipelineRunCmd.Flags().IntVar(&pipelineID, "pipelineid", 0, "pipeline id to retrieve")
	listPipelineRunCmd.MarkFlagRequired("pipelineid")

	getPipelineRunCmd.Flags().IntVar(&pipelineID, "pipelineid", 0, "pipeline id to retrieve")
	getPipelineRunCmd.MarkFlagRequired("pipelineid")

	pipelinesCmd.AddCommand(listPipelinesCmd)
	pipelinesCmd.AddCommand(getPipelineCmd)
	pipelinesCmd.AddCommand(listPipelineRunCmd)
	pipelinesCmd.AddCommand(getPipelineRunCmd)
}

var pipelinesCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Manage an azure devops pipeline",
}

var listPipelinesCmd = &cobra.Command{
	Use:   "list",
	Short: "Get all the pipelines that the project has",
	RunE: func(cmd *cobra.Command, args []string) error {
		if personalAccessToken == "" {
			value, exists := os.LookupEnv("token")
			if exists == false {
				return fmt.Errorf("Please set personal access token or specify it in command line")
			}
			personalAccessToken = value
		}

		organizationURL := fmt.Sprintf("https://dev.azure.com/%s", organization)

		connection := vsts.NewPatConnection(organizationURL, personalAccessToken)

		ctx := context.Background()

		// Create a client to interact with the Pipelines area
		pipelineClient := vstspipelines.NewClient(ctx, connection)

		responseValue, err := pipelineClient.ListPipelines(ctx, vstspipelines.ListPipelinesArgs{
			Project: &project,
		})

		if err != nil {
			fmt.Printf("Get pipelines failed for project %s, error: %v", project, err)
			return fmt.Errorf("Get pipelines failed for project %s, error: %v", project, err)
		}

		index := 0
		if responseValue != nil {
			for _, pipelineReference := range (*responseValue).Value {
				fmt.Printf("Id = %v, Name = %v\n", *pipelineReference.Id, *pipelineReference.Name)
				index++
			}
		}

		return nil
	},
}

var getPipelineCmd = &cobra.Command{
	Use:           "get [pipelineID]",
	Short:         "Get detailed information of the specified pipeline",
	Args:          cobra.ExactArgs(1),
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePipelineID, err := strconv.ParseInt(args[0], 10, 64)
		pipelineID = int(sourcePipelineID)

		if personalAccessToken == "" {
			value, exists := os.LookupEnv("token")
			if exists == false {
				return fmt.Errorf("Please set personal access token or specify it in command line")
			}
			personalAccessToken = value
		}

		organizationURL := fmt.Sprintf("https://dev.azure.com/%s", organization)

		connection := vsts.NewPatConnection(organizationURL, personalAccessToken)

		ctx := context.Background()

		// Create a client to interact with the Pipelines area
		pipelineClient := vstspipelines.NewClient(ctx, connection)

		responseValue, err := pipelineClient.GetPipeline(ctx, vstspipelines.GetPipelineArgs{
			Project:    &project,
			PipelineId: &pipelineID,
		})

		if err != nil {
			fmt.Printf("Get pipeline %d failed for project %s, error: %v", pipelineID, project, err)
			return fmt.Errorf("Get pipeline %d failed for project %s, error: %v", pipelineID, project, err)
		}

		pipeline, err := json.Marshal(responseValue)
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		fmt.Printf("%v", string(pipeline))

		return nil
	},
}

var listPipelineRunCmd = &cobra.Command{
	Use:   "list-run",
	Short: "Get all the runs the specified pipeline triggered",
	RunE: func(cmd *cobra.Command, args []string) error {

		if personalAccessToken == "" {
			value, exists := os.LookupEnv("token")
			if exists == false {
				return fmt.Errorf("Please set personal access token or specify it in command line")
			}
			personalAccessToken = value
		}

		organizationURL := fmt.Sprintf("https://dev.azure.com/%s", organization)

		connection := vsts.NewPatConnection(organizationURL, personalAccessToken)

		ctx := context.Background()

		// Create a client to interact with the Pipelines area
		pipelineClient := vstspipelines.NewClient(ctx, connection)

		responseValue, err := pipelineClient.ListRuns(ctx, vstspipelines.ListRunsArgs{
			Project:    &project,
			PipelineId: &pipelineID,
		})

		if err != nil {
			fmt.Printf("Get pipeline runs failed for pipeline %d, error: %v", pipelineID, err)
			return fmt.Errorf("Get pipeline runs failed for pipeline %d, error: %v", pipelineID, err)
		}

		// index := 0
		if responseValue != nil {
			fmt.Printf("Count = %v\n", len(*responseValue))
		}

		return nil
	},
}

var getPipelineRunCmd = &cobra.Command{
	Use:           "get-run",
	Short:         "Get detailed information of the specified pipeline run",
	Args:          cobra.ExactArgs(1),
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceRunID, err := strconv.ParseInt(args[0], 10, 64)
		runID = int(sourceRunID)

		if personalAccessToken == "" {
			value, exists := os.LookupEnv("token")
			if exists == false {
				return fmt.Errorf("Please set personal access token or specify it in command line")
			}
			personalAccessToken = value
		}

		organizationURL := fmt.Sprintf("https://dev.azure.com/%s", organization)

		connection := vsts.NewPatConnection(organizationURL, personalAccessToken)

		ctx := context.Background()

		// Create a client to interact with the Pipelines area
		pipelineClient := vstspipelines.NewClient(ctx, connection)

		responseValue, err := pipelineClient.GetRun(ctx, vstspipelines.GetRunArgs{
			Project:    &project,
			PipelineId: &pipelineID,
			RunId:      &runID,
		})

		if err != nil {
			fmt.Printf("Get pipeline run %d failed for pipeline %d, error: %v", runID, pipelineID, err)
			return fmt.Errorf("Get pipeline run %d failed for pipeline %d, error: %v", runID, pipelineID, err)
		}

		run, err := json.Marshal(responseValue)
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		fmt.Printf("%v", string(run))

		return nil
	},
}
