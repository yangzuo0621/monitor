package monitor

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/cicd"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/pipelines"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/storageaccountv2"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/vstspat"
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

func (c *MonitorClient) GetDataFromBlob(ctx context.Context, blobName string) (*cicd.Data, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "getDataFromBlob",
		"blob":   blobName,
	})
	blobClient := storageaccountv2.BuildBlobClient(c.azureStorageAccount, c.azureStorageContainer, c.storageAccessKey)
	exist := blobClient.BlobExists(ctx, blobName)

	data := &cicd.Data{}
	logger.Infoln("exist=", exist)
	if exist {
		blob, err := blobClient.GetBlob(ctx, blobName)
		if err != nil {
			logger.WithError(err).Error()
			return nil, err
		}
		err = json.Unmarshal(blob, data)
		if err != nil {
			logger.WithError(err).Error()
			return nil, err
		}
	} else {
		pipelineClient, err := pipelines.BuildPipelineClient(logger, vstspat.NewPATEnvBackend("PERSONAL_ACCESS_TOKEN"), c.organization, c.project)
		if err != nil {
			logger.WithError(err).Error()
			return nil, err
		}

		data.MasterValidation = &cicd.MasterValidation{
			ID: c.masterValidationE2EID,
		}
		data.State = cicd.DataStateValues.None

		builds, err := pipelineClient.ListPipelineBuilds(ctx, c.masterValidationE2EID)
		if len(builds) > 0 {
			build := builds[0]
			logger.Infoln("================== Build ==================")
			bs, _ := json.MarshalIndent(build, "", " ")
			logger.Infoln(string(bs))
			variables := make(map[string]string)
			result, err := pipelineClient.QueueBuild(ctx, 68881, *build.SourceBranch, *build.SourceVersion, variables)
			if err != nil {
				logger.Errorln(err)
				return nil, err
			}

			logger.Infoln("================== Result ==================")
			bs, _ = json.MarshalIndent(result, "", " ")
			logger.Infoln(string(bs))

			// "vstfs:///Build/Build/34898972"
			ss := strings.Split(*result.Uri, "/")
			id := ss[len(ss)-1]
			i, _ := strconv.ParseInt(id, 10, 64)

			data.MasterValidation.Branch = *build.SourceBranch
			data.MasterValidation.CommitID = *build.SourceVersion
			data.AKSBuild = &cicd.AKSBuild{
				ID:    int(i),
				Count: 0,
			}
			data.State = cicd.DataStateValues.NotStart
		}

		content, _ := json.Marshal(data)
		r, err := blobClient.UploadBlob(ctx, blobName, content)
		if err != nil {
			logger.Errorln(err)
			return nil, err
		}
		logger.Infoln("================== Status ==================")
		logger.Infoln("status code=", r)
	}

	return data, nil
}
