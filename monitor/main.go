package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/akv"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/cicd"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/pipelines"
	"github.com/yangzuo0621/azure-devops-cmd/monitor/pkg/storageaccountv2"
)

const (
	clientID           = ""
	tenantID           = ""
	vaultName          = "zuya20200920akv"
	clientSecret       = ""
	accountName        = "zuya20200920account"
	containerName      = "test1"
	accessKeyName      = "AccessKey"
	organization       = "msazure"
	project            = "CloudNativeCompute"
	masterValidationID = 138746
	devAKSDeployID     = 68881
	aksBuildID         = 74751
)

const gitRepoFormat = "https://dev.azure.com/%s/%s/_git/%s"

func checkErr(err error) {
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(-1)
	}
}

func main() {
	now := time.Now().UTC()
	date := now.Format("2006-01-02")
	fmt.Println("date=", date)
	ctx := context.Background()
	akvClient := akv.BuildAKVClient(clientID, tenantID, clientSecret, vaultName)

	accessKey, err := akvClient.GetSecretFromAzureKeyVault(ctx, accessKeyName)
	checkErr(err)

	pipelineClient, err := pipelines.BuildPipelineClient(logrus.New(), akvClient, organization, project)
	checkErr(err)

	blobClient := storageaccountv2.BuildBlobClient(accountName, containerName, *accessKey)
	exist := blobClient.BlobExists(ctx, date)

	if exist {
		for true {
			fmt.Println("exist")
			blob, err := blobClient.GetBlob(ctx, date)
			checkErr(err)
			data := cicd.Data{}
			json.Unmarshal(blob, &data)
			fmt.Println(data)

			result, err := pipelineClient.GetPipelineBuildByID(ctx, data.BuildID)
			checkErr(err)

			data.BuildStatus = string(*result.Status)
			content, _ := json.Marshal(data)
			r, err := blobClient.UploadBlob(ctx, date, content)
			checkErr(err)
			fmt.Println("status code=", r)

			if strings.EqualFold(string(*result.Status), "Completed") {
				fmt.Println("Completed")
				break
			} else {
				fmt.Println("Sleep 5 mins")
				time.Sleep(5 * time.Minute)
			}
		}
	} else {
		builds, err := pipelineClient.ListPipelineBuilds(ctx, masterValidationID)
		checkErr(err)
		for _, b := range builds {
			fmt.Printf("%-11s %d %s\n", *b.BuildNumber, *b.Id, *b.Result)
		}

		if len(builds) > 0 {
			build := builds[0]
			fmt.Println("================== Build ==================")
			bs, _ := json.MarshalIndent(build, "", " ")
			fmt.Println(string(bs))
			variables := make(map[string]string)
			// variables["AKS_E2E_UNDERLAY_TYPE"] = "AKS_ENGINE_CLUSTER"
			result, err := pipelineClient.QueueBuild(ctx, aksBuildID, *build.SourceBranch, *build.SourceVersion, variables)
			checkErr(err)

			fmt.Println("================== Result ==================")
			bs, _ = json.MarshalIndent(result, "", " ")
			fmt.Println(string(bs))

			// "vstfs:///Build/Build/34898972"
			ss := strings.Split(*result.Uri, "/")
			id := ss[len(ss)-1]
			i, _ := strconv.ParseInt(id, 10, 64)

			data := cicd.Data{
				BuildID:  int(i),
				CommitID: *build.SourceVersion,
			}

			content, _ := json.Marshal(data)
			r, err := blobClient.UploadBlob(ctx, date, content)
			checkErr(err)
			fmt.Println("================== Status ==================")
			fmt.Println("status code=", r)
		}
	}
}
