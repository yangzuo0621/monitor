package pipeline

import "github.com/spf13/cobra"

// CreateCommand creates a cobra command instance fot pipeline.
func CreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Manage an azure DevOps pipeline",
	}

	return cmd
}

func createListPipelineCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "Get all the pipelines that the project has",
	}

	return c
}
