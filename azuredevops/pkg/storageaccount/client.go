package storageaccount

import (
	"context"
	"fmt"
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

func (c *Client) GetAccountKeys() (*storage.AccountListKeysResult, error) {
	accountsClient, err := c.GetStorageAccountClient()
	if err != nil {
		return nil, err
	}
	result, err := accountsClient.ListKeys(context.Background(), c.ResourceGroup, c.AccountName)
	return &result, err
}

func (c *Client) GetAccountPrimaryKey() string {
	response, err := c.GetAccountKeys()
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*response.Keys)[0]).Value)
}

func (c *Client) GetContainerURL(containerName string) azblob.ContainerURL {
	key := c.GetAccountPrimaryKey()
	cred, _ := azblob.NewSharedKeyCredential(c.AccountName, key)
	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, c.AccountName))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

func (c *Client) GetContainer(containerName string) (azblob.ContainerURL, error) {
	container := c.GetContainerURL(containerName)

	_, err := container.GetProperties(context.Background(), azblob.LeaseAccessConditions{})
	return container, err
}

func (c *Client) CreateContainer(containerName string) (azblob.ContainerURL, error) {
	container := c.GetContainerURL(containerName)

	_, err := container.Create(
		context.Background(),
		azblob.Metadata{},
		azblob.PublicAccessContainer)
	return container, err
}
