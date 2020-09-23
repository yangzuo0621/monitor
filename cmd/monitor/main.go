package main

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/monitor/pkg/monitor"
)

const (
	organizationKey          = "ORGANIZATION"
	projectKey               = "PROJECT"
	masterValidationE2EIDKey = "MASTER_VALIDATION_E2E_ID"
	aksBuildIDKey            = "AKS_BUILD_ID"
	releaseIDKey             = "RELEASE_ID"
	azureStorageAccountKey   = "AZURE_STORAGE_ACCOUNT"
	azureStorageContainerKey = "AZURE_STORAGE_CONTAINER"
	storageAccessKeyKey      = "AZURE_STORAGE_ACCESS_KEY"
	personalAccessTokenKey   = "PERSONAL_ACCESS_TOKEN"
	monitorTimeInterval      = 5
)

var (
	organization          string
	project               string
	masterValidationE2EID int
	aksBuildID            int
	releaseID             int
	azureStorageAccount   string
	azureStorageContainer string
	storageAccessKey      string
	personalAccessToken   string

	logger *logrus.Entry
)

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"source": "monitor",
	})

	organization = os.Getenv(organizationKey)
	project = os.Getenv(projectKey)
	i, _ := strconv.ParseInt(os.Getenv(masterValidationE2EIDKey), 10, 64)
	masterValidationE2EID = int(i)

	i, _ = strconv.ParseInt(os.Getenv(aksBuildIDKey), 0, 64)
	aksBuildID = int(i)

	i, _ = strconv.ParseInt(os.Getenv(releaseIDKey), 0, 64)
	releaseID = int(i)

	azureStorageAccount = os.Getenv(azureStorageAccountKey)
	azureStorageContainer = os.Getenv(azureStorageContainerKey)
	storageAccessKey = os.Getenv(storageAccessKeyKey)
	personalAccessToken = os.Getenv(personalAccessTokenKey)
}

func main() {
	client := monitor.BuildClient(
		organization,
		project,
		masterValidationE2EID,
		aksBuildID,
		releaseID,
		azureStorageAccount,
		azureStorageContainer,
		storageAccessKey,
		personalAccessToken,
		logger,
	)

	client.MonitorRoutine()
}
