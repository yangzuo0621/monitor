package main

import "github.com/spf13/cobra"

// RootCmd creates a cobra command instance of azuredevops.
func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "azuredevops",
		Short: "A command line tool for operate Azure DevOps Service",
	}

	return rootCmd
}
