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

	logger.Infof(
		"organization=%s, project=%s, masterValidationE2EID=%d, aksBuildID=%d, releaseID=%d, azureStorageAccount=%s, azureStorageContainer=%s, storageAccessKey=%s, personalAccessToken=%s",
		organization,
		project,
		masterValidationE2EID,
		aksBuildID,
		releaseID,
		azureStorageAccount,
		azureStorageContainer,
		storageAccessKey,
		personalAccessToken,
	)
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
		logger.Infoln(cicd.DataStateValues.None)
	case cicd.DataStateValues.NotStart:
		logger.Infoln(cicd.DataStateValues.NotStart)
	default:
		logger.Infoln("default")
	}

	// pipelineClient, err := pipelines.BuildPipelineClient(logger, vstspat.NewPATEnvBackend(personalAccessTokenKey), organization, project)
	// if err != nil {
	// 	logger.Errorln(err)
	// 	return
	// }

	// blobClient := storageaccountv2.BuildBlobClient(azureStorageAccount, azureStorageContainer, storageAccessKey)
	// exist := blobClient.BlobExists(ctx, date)

	// if exist {
	// 	logger.Infoln("exist")
	// 	blob, err := blobClient.GetBlob(ctx, date)
	// 	if err != nil {
	// 		logger.Errorln(err)
	// 		return
	// 	}
	// 	data := cicd.Data{}
	// 	json.Unmarshal(blob, &data)
	// 	logger.Infoln(data)

	// 	result, err := pipelineClient.GetPipelineBuildByID(ctx, data.BuildID)
	// 	if err != nil {
	// 		logger.Errorln(err)
	// 		return
	// 	}

	// 	data.BuildStatus = string(*result.Status)
	// 	content, _ := json.Marshal(data)
	// 	r, err := blobClient.UploadBlob(ctx, date, content)
	// 	if err != nil {
	// 		logger.Errorln(err)
	// 		return
	// 	}
	// 	logger.Infoln("status code=", r)

	// 	if strings.EqualFold(string(*result.Status), "Completed") {
	// 		logger.Infoln("Completed")
	// 	}
	// } else {
	// 	builds, err := pipelineClient.ListPipelineBuilds(ctx, masterValidationE2EID)
	// 	if err != nil {
	// 		logger.Errorln(err)
	// 		return
	// 	}
	// 	for _, b := range builds {
	// 		logger.Infof("%-11s %d %s\n", *b.BuildNumber, *b.Id, *b.Result)
	// 	}

	// 	if len(builds) > 0 {
	// 		build := builds[0]
	// 		logger.Infoln("================== Build ==================")
	// 		bs, _ := json.MarshalIndent(build, "", " ")
	// 		logger.Infoln(string(bs))
	// 		variables := make(map[string]string)
	// 		// variables["AKS_E2E_UNDERLAY_TYPE"] = "AKS_ENGINE_CLUSTER"
	// 		// result, err := pipelineClient.QueueBuild(ctx, aksBuildID, *build.SourceBranch, *build.SourceVersion, variables)
	// 		result, err := pipelineClient.QueueBuild(ctx, 68881, *build.SourceBranch, *build.SourceVersion, variables)
	// 		if err != nil {
	// 			logger.Errorln(err)
	// 			return
	// 		}

	// 		fmt.Println("================== Result ==================")
	// 		bs, _ = json.MarshalIndent(result, "", " ")
	// 		fmt.Println(string(bs))

	// 		// "vstfs:///Build/Build/34898972"
	// 		ss := strings.Split(*result.Uri, "/")
	// 		id := ss[len(ss)-1]
	// 		i, _ := strconv.ParseInt(id, 10, 64)

	// 		data := cicd.Data{
	// 			BuildID:  int(i),
	// 			CommitID: *build.SourceVersion,
	// 		}

	// 		content, _ := json.Marshal(data)
	// 		r, err := blobClient.UploadBlob(ctx, date, content)
	// 		if err != nil {
	// 			logger.Errorln(err)
	// 			return
	// 		}
	// 		logger.Infoln("================== Status ==================")
	// 		logger.Infoln("status code=", r)
	// 	}
	// }
}
