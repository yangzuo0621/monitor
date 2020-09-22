package monitor

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	vstsbuild "github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/azure-devops-cmd/pkg/cicd"
	"github.com/yangzuo0621/azure-devops-cmd/pkg/pipelines"
	"github.com/yangzuo0621/azure-devops-cmd/pkg/storageaccountv2"
	"github.com/yangzuo0621/azure-devops-cmd/pkg/vstspat"
)

const (
	personalAccessTokenKey = "PERSONAL_ACCESS_TOKEN"
)

type MonitorClient struct {
	organization          string
	project               string
	masterValidationE2EID int
	aksBuildID            int
	releaseID             int
	azureStorageAccount   string
	azureStorageContainer string
	storageAccessKey      string
	personalAccessToken   string

	logger logrus.FieldLogger
}

func BuildClient(
	organization string,
	project string,
	masterValidationE2EID int,
	aksBuildID int,
	releaseID int,
	azureStorageAccount string,
	azureStorageContainer string,
	storageAccessKey string,
	personalAccessToken string,
	rootLogger logrus.FieldLogger,
) *MonitorClient {
	logger := rootLogger.WithFields(logrus.Fields{
		"organization": organization,
		"project":      project,
	})
	return &MonitorClient{
		organization:          organization,
		project:               project,
		masterValidationE2EID: masterValidationE2EID,
		aksBuildID:            aksBuildID,
		releaseID:             releaseID,
		azureStorageAccount:   azureStorageAccount,
		azureStorageContainer: azureStorageContainer,
		storageAccessKey:      storageAccessKey,
		personalAccessToken:   personalAccessToken,
		logger:                logger,
	}
}

// GetDataFromBlob retrives data from azure storage account blob
func (c *MonitorClient) GetDataFromBlob(ctx context.Context, blobName string) (*cicd.Data, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "getDataFromBlob",
		"blob":   blobName,
	})

	blobClient := storageaccountv2.BuildBlobClient(c.azureStorageAccount, c.azureStorageContainer, c.storageAccessKey)
	exist := blobClient.BlobExists(ctx, blobName)

	var data cicd.Data
	if exist {
		blob, err := blobClient.GetBlob(ctx, blobName)
		if err != nil {
			logger.WithError(err).Error()
			return nil, err
		}
		err = json.Unmarshal(blob, &data)
		if err != nil {
			logger.WithError(err).Error()
			return nil, err
		}
	} else {
		data = cicd.Data{
			MasterValidation: &cicd.MasterValidation{
				ID: c.masterValidationE2EID,
			},
			State: cicd.DataStateValues.None,
		}
	}

	return &data, nil
}

// UploadDataToBlob update data to azure storage account blob
func (c *MonitorClient) UploadDataToBlob(ctx context.Context, blobName string, data *cicd.Data) error {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "UploadDataToBlob",
		"blob":   blobName,
	})

	blobClient := storageaccountv2.BuildBlobClient(c.azureStorageAccount, c.azureStorageContainer, c.storageAccessKey)
	content, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	_, err = blobClient.UploadBlob(ctx, blobName, content)
	if err != nil {
		logger.WithError(err).Error()
	}
	return err
}

// TriggerAKSBuild triggers [EV2] AKS Build
func (c *MonitorClient) TriggerAKSBuild(ctx context.Context, data *cicd.Data) error {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "TriggerAKSBuild",
	})

	pipelineClient, err := pipelines.BuildPipelineClient(logger, vstspat.NewPATEnvBackend(personalAccessTokenKey), c.organization, c.project)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	builds, err := pipelineClient.ListPipelineBuilds(ctx, c.masterValidationE2EID)
	if len(builds) > 0 {
		build := builds[0]
		logger.Infoln("================== Build ==================")
		bs, _ := json.MarshalIndent(build, "", " ")
		logger.Infoln(string(bs))
		variables := make(map[string]string)
		// result, err := pipelineClient.QueueBuild(ctx, c.aksBuildID, *build.SourceBranch, *build.SourceVersion, variables)
		// 68881 just for testing
		result, err := pipelineClient.QueueBuild(ctx, 68881, *build.SourceBranch, *build.SourceVersion, variables)
		if err != nil {
			logger.Errorln(err)
			return err
		}

		logger.Infoln("================== Result ==================")
		bs, _ = json.MarshalIndent(result, "", " ")
		logger.Infoln(string(bs))

		// "vstfs:///Build/Build/34898972"
		ss := strings.Split(*result.Uri, "/")
		id := ss[len(ss)-1]
		i, _ := strconv.ParseInt(id, 10, 64)

		data.MasterValidation.Branch = build.SourceBranch
		data.MasterValidation.CommitID = build.SourceVersion

		if data.AKSBuild != nil {
			data.AKSBuild.ID = int(i)
			data.AKSBuild.Count = data.AKSBuild.Count + 1
		} else {
			data.AKSBuild = &cicd.AKSBuild{
				ID:    int(i),
				Count: 1,
			}
		}

		data.State = cicd.DataStateValues.NotStart
	}
	return nil
}

// MonitorAKSBuild monitors the running status of [EV2] AKS Build
func (c *MonitorClient) MonitorAKSBuild(ctx context.Context, data *cicd.Data) error {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "MonitorAKSBuild",
	})

	pipelineClient, err := pipelines.BuildPipelineClient(logger, vstspat.NewPATEnvBackend(personalAccessTokenKey), c.organization, c.project)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	build, err := pipelineClient.GetPipelineBuildByID(ctx, data.AKSBuild.ID)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	status := string(*build.Status)

	if *build.Status == vstsbuild.BuildStatusValues.Completed {
		if *build.Result == vstsbuild.BuildResultValues.Succeeded {
			data.State = cicd.DataStateValues.BuildSucceeded
		} else {
			data.State = cicd.DataStateValues.BuildFailed
		}
		result := string(*build.Result)
		data.AKSBuild.BuildResult = &result
	} else {
		data.State = cicd.DataStateValues.BuildInProgress
	}
	data.AKSBuild.BuildStatus = &status
	return nil
}
