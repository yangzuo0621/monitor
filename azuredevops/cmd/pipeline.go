package cmd

import (
	"context"
	"fmt"
	"os"

	vsts "github.com/microsoft/azure-devops-go-api/azuredevops"
	vstspipelines "github.com/microsoft/azure-devops-go-api/azuredevops/pipelines"
	"github.com/spf13/cobra"
)

var (
	organization        string
	project             string
	author              string
	personalAccessToken string
)

func init() {
	pipelinesCmd.PersistentFlags().StringVarP(&organization, "organization", "o", "", "Organization the pipeline belongs to (required)")
	pipelinesCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project the pipeline belongs to (required)")
	pipelinesCmd.PersistentFlags().StringVarP(&personalAccessToken, "token", "t", "", "Personal access token to azure devops")
	pipelinesCmd.MarkPersistentFlagRequired("organization")
	pipelinesCmd.MarkPersistentFlagRequired("project")

	pipelinesCmd.AddCommand(listPipelinesCmd)
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
