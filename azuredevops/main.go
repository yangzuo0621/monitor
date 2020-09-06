package main

import (
	"os"

	"github.com/yangzuo0621/azure-devops-cmd/azuredevops/cmd"
)

func main() {
	rootCmd := cmd.RootCmd()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
