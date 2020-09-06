package cmd

import "github.com/spf13/cobra"

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "azuredevops",
		Short: "A command line tool for operate Azure DevOps Service",
	}

	rootCmd.AddCommand(pipelinesCmd)

	return rootCmd
}
