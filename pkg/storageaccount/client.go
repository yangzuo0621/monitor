package storageaccount

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2018-07-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type Client struct {
	ClientID       string
	TenantID       string
	ClientSecret   string
	ResourceGroup  string
	AccountName    string
	SubscriptionID string
}

var (
	cloudName        string = "AzurePublicCloud"
	environment      *azure.Environment
	blobFormatString = `https://%s.blob.core.windows.net`
)

func Environment() (*azure.Environment, error) {
	if environment != nil {
		return environment, nil
	}

	env, err := azure.EnvironmentFromName(cloudName)
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// GetAzureKeyVaultAuthorizer constructs authorizer for storage account
func (c *Client) GetAzureStorageAccountAuthorizer() (autorest.Authorizer, error) {
	cloudEnv, err := Environment()
	if err != nil {
		return nil, err
	}

	oauthConfig, err := adal.NewOAuthConfig(cloudEnv.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		return nil, fmt.Errorf("create OAuth config failed")
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, cloudEnv.ResourceManagerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("create service principal token failed: %w", err)
	}

	return autorest.NewBearerAuthorizer(token), nil
}

// GetStorageAccountClient creates a storage account client
func (c *Client) GetStorageAccountClient() (*storage.AccountsClient, error) {
	storageAccountsClient := storage.NewAccountsClient(c.SubscriptionID)
	auth, err := c.GetAzureStorageAccountAuthorizer()
	if err != nil {
		return nil, err
	}
	storageAccountsClient.Authorizer = auth
	return &storageAccountsClient, nil
}

func (c *Client) GetAccountKeys(ctx context.Context) (*storage.AccountListKeysResult, error) {
	accountsClient, err := c.GetStorageAccountClient()
	if err != nil {
		return nil, err
	}
	result, err := accountsClient.ListKeys(ctx, c.ResourceGroup, c.AccountName)
	return &result, err
}

func (c *Client) GetAccountPrimaryKey(ctx context.Context) string {
	response, err := c.GetAccountKeys(ctx)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*response.Keys)[0]).Value)
}

func (c *Client) GetContainerURL(ctx context.Context, containerName string) azblob.ContainerURL {
	key := c.GetAccountPrimaryKey(ctx)
	cred, _ := azblob.NewSharedKeyCredential(c.AccountName, key)
	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, c.AccountName))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

func (c *Client) GetContainerURLFromAccessKey(ctx context.Context, containerName string, accessKey string) azblob.ContainerURL {
	cred, _ := azblob.NewSharedKeyCredential(c.AccountName, accessKey)
	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, c.AccountName))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

func (c *Client) GetContainer(ctx context.Context, containerName string) (azblob.ContainerURL, error) {
	container := c.GetContainerURL(ctx, containerName)

	_, err := container.GetProperties(ctx, azblob.LeaseAccessConditions{})
	return container, err
}

func (c *Client) CreateContainer(ctx context.Context, containerName string) (azblob.ContainerURL, error) {
	container := c.GetContainerURL(ctx, containerName)

	_, err := container.Create(
		ctx,
		azblob.Metadata{},
		azblob.PublicAccessContainer)
	return container, err
}

func (c *Client) getBlobURL(ctx context.Context, containerName string, blobName string) azblob.BlobURL {
	container := c.GetContainerURL(ctx, containerName)
	blob := container.NewBlobURL(blobName)
	return blob
}

func (c *Client) GetBlob(ctx context.Context, containerName string, blobName string) (string, error) {
	b := c.getBlobURL(ctx, containerName, blobName)

	resp, err := b.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return "", err
	}
	defer resp.Response().Body.Close()
	body, err := ioutil.ReadAll(resp.Body(azblob.RetryReaderOptions{}))
	return string(body), err
}

func (c *Client) UploadBlob(ctx context.Context, containerName string, blobName string, content string) (int, error) {
	b := c.getBlobURL(ctx, containerName, blobName)

	resp, err := b.ToBlockBlobURL().Upload(ctx, bytes.NewReader([]byte(content)), azblob.BlobHTTPHeaders{}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	if err != nil {
		return 0, err
	}

	return resp.StatusCode(), nil
}
