package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/cicd"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/monitor"
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
	monitorTimeInterval      = 1
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
	for true {
		select {
		case <-time.After(monitorTimeInterval * time.Minute):
			run()
		}
	}
}

func run() {
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

	now := time.Now().UTC()
	date := now.Format("2006-01-02")
	logger.Infoln("date=", date)
	ctx := context.Background()
	data, err := client.GetDataFromBlob(ctx, date)
	if err != nil {
		logger.Errorln(err)
		return
	}
	logger.Infof("%v", data)

	switch data.State {
	case cicd.DataStateValues.None:
		client.TriggerAKSBuild(ctx, data)
	case cicd.DataStateValues.NotStart, cicd.DataStateValues.BuildInProgress:
		client.MonitorAKSBuild(ctx, data)
	case cicd.DataStateValues.BuildFailed:
		client.TriggerAKSBuild(ctx, data)
	case cicd.DataStateValues.BuildSucceeded:
		logger.Infoln("Trigger release pipeline")
	default:
		logger.Infoln("default")
	}

	err = client.UploadDataToBlob(ctx, date, data)
	if err != nil {
		logger.Errorln(err)
		return
	}
}
