package main

import (
	"os"

	"github.com/yangzuo0621/azure-devops-cmd/azuredevops/pkg/pipelines"
)

func main() {
	rootCmd := RootCmd()

	rootCmd.AddCommand(pipelines.CreateCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
