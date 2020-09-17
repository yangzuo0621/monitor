package akv

import (
	"context"

	"github.com/Azure/go-autorest/autorest"
)

// AKVClient interface for manuplate Azure Keyvault
type AKVClient interface {
	// GetAzureKeyVaultAuthorizer return azure keyvault authorizer
	GetAzureKeyVaultAuthorizer() (autorest.Authorizer, error)

	// GetSecretFromAzureKeyVault return the secret value of secret name from AKV
	GetSecretFromAzureKeyVault(ctx context.Context, secretName string) (*string, error)

	// GetPAT gets a personal access token from akv
	GetPAT(ctx context.Context) (string, error)
}
