package storageaccountv2

import (
	"context"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

// BlobClient interface for manuplate Blob of storage account container
type BlobClient interface {
	// GetContainerURL return destination Container
	GetContainerURL() (*azblob.ContainerURL, error)

	// BlobExists check whether the blob exists
	BlobExists(ctx context.Context, blobName string) bool

	// GetBlob get the contents of blob
	GetBlob(ctx context.Context, blobName string) ([]byte, error)

	// UploadBlob create or update blob content
	UploadBlob(ctx context.Context, blobName string, content []byte) (int, error)
}
