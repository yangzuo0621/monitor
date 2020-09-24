package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yangzuo0621/monitor/pkg/monitor"
)

const (
	storageAccessKeyKey    = "AZURE_STORAGE_ACCESS_KEY"
	personalAccessTokenKey = "PERSONAL_ACCESS_TOKEN"
)

var (
	storageAccessKey    string
	personalAccessToken string
	configPath          string

	logger *logrus.Entry
)

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"source": "monitor",
	})

	storageAccessKey = os.Getenv(storageAccessKeyKey)
	if storageAccessKey == "" {
		logger.Fatalln("env storageAccessKey not set")
		os.Exit(-1)
	}
	personalAccessToken = os.Getenv(personalAccessTokenKey)
	if personalAccessToken == "" {
		logger.Fatalln("env personalAccessToken not set")
		os.Exit(-1)
	}
}

func main() {
	rootCmd := &cobra.Command{
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

	rootCmd.Flags().StringVar(&configPath, "config", "", "config file path")
	rootCmd.MarkFlagRequired("config")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}

}
