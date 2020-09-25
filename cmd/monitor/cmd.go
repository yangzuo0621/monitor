package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/yangzuo0621/monitor/pkg/monitor"
)

func createRootCmd() *cobra.Command {
	var configPath string

	c := &cobra.Command{
		Use:          "monitor",
		Short:        "monitor CI/CD process",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configContent, err := ioutil.ReadFile(configPath)
			if err != nil {
				return err
			}
			var c monitor.Config
			err = json.Unmarshal(configContent, &c)
			if err != nil {
				return err
			}

			client := monitor.BuildClient(
				storageAccessKey,
				personalAccessToken,
				&c,
				logger,
			)

			client.MonitorRoutine()
			return nil
		},
	}

	c.Flags().StringVar(&configPath, "config", "", "config file path")
	c.MarkFlagRequired("config")

	return c
}
