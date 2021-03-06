package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	vstsbuild "github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/monitor/pkg/cicd"
	"github.com/yangzuo0621/monitor/pkg/pipelines"
	"github.com/yangzuo0621/monitor/pkg/releases"
	"github.com/yangzuo0621/monitor/pkg/storageaccountv2"
	"github.com/yangzuo0621/monitor/pkg/vstspat"
)

const (
	personalAccessTokenKey = "PERSONAL_ACCESS_TOKEN"
	monitorTimeInterval    = 5
)

// MonitorClient encapsulates the data client needs
type MonitorClient struct {
	storageAccessKey    string
	personalAccessToken string
	config              *Config

	logger logrus.FieldLogger
}

type ReleaseMetadata struct {
	ID       int
	Alias    string
	Stagings []string
}

type Config struct {
	Organization          string     `json:"organization"`
	Project               string     `json:"project"`
	MasterValidationE2EID int        `json:"master_validation_e2e_id"`
	AksBuildID            int        `json:"aks_build_id"`
	AksRelease            []*Release `json:"aks_release"`
	AzureStorageAccount   string     `json:"azure_storage_account"`
	AzureStorageContainer string     `json:"azure_storage_container"`
}

type Release struct {
	DefinitionID int      `json:"definition_id"`
	Alias        string   `json:"source_alias"`
	Stagings     []string `json:"staging"`
}

// BuildClient creates an instance of MonitorClient
func BuildClient(
	storageAccessKey string,
	personalAccessToken string,
	config *Config,
	rootLogger logrus.FieldLogger,
) *MonitorClient {
	logger := rootLogger.WithFields(logrus.Fields{
		"organization": config.Organization,
		"project":      config.Project,
	})
	return &MonitorClient{
		storageAccessKey:    storageAccessKey,
		personalAccessToken: personalAccessToken,
		config:              config,
		logger:              logger,
	}
}

// MonitorRoutine monitors the status of CI/CD every 5 minutes
func (c *MonitorClient) MonitorRoutine() {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "MonitorRoutine",
	})

	for true {
		select {
		case <-time.After(monitorTimeInterval * time.Minute):
			now := time.Now().UTC()
			date := now.Format("2006-01-02")
			logger.Infoln("date=", date)
			ctx := context.Background()
			data, err := c.GetDataFromBlob(ctx, date)
			if err != nil {
				logger.Errorln(err)
				return
			}
			logger.Infof("%v", data)

			switch data.State {
			case cicd.DataStateValues.None:
				c.TriggerAKSBuild(ctx, data)
			case cicd.DataStateValues.NotStart, cicd.DataStateValues.BuildInProgress:
				c.MonitorAKSBuild(ctx, data)
			case cicd.DataStateValues.BuildFailed:
				c.TriggerAKSBuild(ctx, data)
			case cicd.DataStateValues.BuildSucceeded:
				c.TriggerRelease(ctx, data)
			case cicd.DataStateValues.ReleaseInProgress:
				c.MonitorRelease(ctx, data)
			default:
				logger.Infoln("default")
			}

			err = c.UploadDataToBlob(ctx, date, data)
			if err != nil {
				logger.Errorln(err)
			}
		}
	}
}

// GetDataFromBlob retrives data from azure storage account blob
func (c *MonitorClient) GetDataFromBlob(ctx context.Context, blobName string) (*cicd.Data, error) {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "getDataFromBlob",
		"blob":   blobName,
	})

	blobClient := storageaccountv2.BuildBlobClient(c.config.AzureStorageAccount, c.config.AzureStorageContainer, c.storageAccessKey, logger)
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
		length := len(c.config.AksRelease)
		aksReleases := make([]*cicd.AKSRelease, length)

		for i := 0; i < length; i++ {
			ss := c.config.AksRelease[i].Stagings
			stagings := make([]*cicd.Staging, len(ss))
			for j := 0; j < len(ss); j++ {
				stagings[j] = &cicd.Staging{
					Name: ss[j],
				}
			}
			aksReleases[i] = &cicd.AKSRelease{
				DefinitionID: c.config.AksRelease[i].DefinitionID,
				Alias:        c.config.AksRelease[i].Alias,
				Staging:      stagings,
			}
		}

		data = cicd.Data{
			MasterValidation: &cicd.MasterValidation{
				ID: c.config.MasterValidationE2EID,
			},
			State:      cicd.DataStateValues.None,
			AKSRelease: aksReleases,
			Date:       blobName,
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

	blobClient := storageaccountv2.BuildBlobClient(c.config.AzureStorageAccount, c.config.AzureStorageContainer, c.storageAccessKey, logger)
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

	pipelineClient, err := pipelines.BuildPipelineClient(logger, vstspat.NewPATEnvBackend(personalAccessTokenKey), c.config.Organization, c.config.Project)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	builds, err := pipelineClient.ListPipelineBuilds(ctx, c.config.MasterValidationE2EID)
	if len(builds) > 0 {
		build := builds[0]
		logger.Infoln("================== Build ==================")
		bs, _ := json.MarshalIndent(build, "", " ")
		logger.Infoln(string(bs))
		variables := make(map[string]string)
		result, err := pipelineClient.QueueBuildByCommit(ctx, c.config.AksBuildID, *build.SourceVersion, variables)
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
			data.AKSBuild.BuildNumber = result.BuildNumber
			data.AKSBuild.BuildResult = nil
			data.AKSBuild.BuildStatus = nil
		} else {
			data.AKSBuild = &cicd.AKSBuild{
				ID:          int(i),
				BuildNumber: result.BuildNumber,
				Count:       1,
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

	pipelineClient, err := pipelines.BuildPipelineClient(logger, vstspat.NewPATEnvBackend(personalAccessTokenKey), c.config.Organization, c.config.Project)
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

// TriggerRelease triggers a release
func (c *MonitorClient) TriggerRelease(ctx context.Context, data *cicd.Data) error {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "TriggerRelease",
	})

	releaseClient, err := releases.BuildReleaseClient(logger, vstspat.NewPATEnvBackend(personalAccessTokenKey), c.config.Organization, c.config.Project)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	var resultErr error = nil
	for _, v := range data.AKSRelease {
		release, err := releaseClient.CreateRelease(ctx, v.DefinitionID, v.Alias, strconv.Itoa(data.AKSBuild.ID), *data.AKSBuild.BuildNumber, fmt.Sprintf("Daily release: %s", data.Date))
		if err != nil {
			logger.WithError(err).Error()
			resultErr = err
		} else {
			v.ReleaseID = release.Id
			v.ReleaseName = release.Name
		}
	}

	data.State = cicd.DataStateValues.ReleaseInProgress
	return resultErr
}

// MonitorRelease monitors the status of release
func (c *MonitorClient) MonitorRelease(ctx context.Context, data *cicd.Data) error {
	logger := c.logger.WithFields(logrus.Fields{
		"action": "MonitorRelease",
	})

	releaseClient, err := releases.BuildReleaseClient(logger, vstspat.NewPATEnvBackend(personalAccessTokenKey), c.config.Organization, c.config.Project)
	if err != nil {
		logger.WithError(err).Error()
		return err
	}

	var resultErr error = nil
	for _, v := range data.AKSRelease {
		release, err := releaseClient.GetReleaseByID(ctx, *v.ReleaseID)
		if err != nil {
			logger.WithError(err).Error()
			resultErr = err
		} else {
			for _, s := range v.Staging {
				for _, e := range *release.Environments {
					if strings.EqualFold(s.Name, *e.Name) {
						status := string(*e.Status)
						s.Status = &status
						break
					}
				}
			}
		}
	}

	return resultErr
}
