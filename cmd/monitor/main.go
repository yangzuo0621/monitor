package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

const (
	storageAccessKeyKey    = "AZURE_STORAGE_ACCESS_KEY"
	personalAccessTokenKey = "PERSONAL_ACCESS_TOKEN"
)

var (
	storageAccessKey    string
	personalAccessToken string

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
	rootCmd := createRootCmd()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}

}
