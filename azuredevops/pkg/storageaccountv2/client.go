package storageaccountv2

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type blobClient struct {
	accountName   string
	containerName string
	accessKey     string
}

const blobURL = `https://%s.blob.core.windows.net`

// BuildBlobClient build a blob client instance
func BuildBlobClient(accountName string, containerName string, accessKey string) BlobClient {
	return &blobClient{
		accountName:   accountName,
		containerName: containerName,
		accessKey:     accessKey,
	}
}

func (c *blobClient) GetContainerURL() (*azblob.ContainerURL, error) {
	cred, err := azblob.NewSharedKeyCredential(c.accountName, c.accessKey)
	if err != nil {
		return nil, err
	}
	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf(blobURL, c.accountName))
	if err != nil {
		return nil, err
	}
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(c.containerName)
	return &container, nil
}

func (c *blobClient) BlobExists(ctx context.Context, blobName string) bool {
	container, err := c.GetContainerURL()
	if err != nil {
		return false
	}
	blob := container.NewBlobURL(blobName)

	_, err = blob.GetProperties(ctx, azblob.BlobAccessConditions{})
	if err != nil {
		return false
	}
	return true
}

func (c *blobClient) GetBlob(ctx context.Context, blobName string) ([]byte, error) {
	container, err := c.GetContainerURL()
	if err != nil {
		return nil, err
	}
	blob := container.NewBlobURL(blobName)

	resp, err := blob.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, err
	}
	defer resp.Response().Body.Close()
	body, err := ioutil.ReadAll(resp.Body(azblob.RetryReaderOptions{}))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *blobClient) UploadBlob(ctx context.Context, blobName string, content []byte) (int, error) {
	container, err := c.GetContainerURL()
	if err != nil {
		return 400, err
	}
	blob := container.NewBlobURL(blobName)

	resp, err := blob.ToBlockBlobURL().Upload(
		ctx,
		bytes.NewReader(content),
		azblob.BlobHTTPHeaders{},
		azblob.Metadata{},
		azblob.BlobAccessConditions{})

	if err != nil {
		return 400, err
	}
	return resp.StatusCode(), nil
}

var _ BlobClient = (*blobClient)(nil)
