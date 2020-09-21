package main

import (
	"os"

	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/git"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/pipelines"
)

func main() {
	rootCmd := RootCmd()

	rootCmd.AddCommand(pipelines.CreateCommand())
	rootCmd.AddCommand(git.CreateCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
