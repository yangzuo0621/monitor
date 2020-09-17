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
	GetSecretFromAzureKeyVault(ctx context.Context, vaultName string, secretName string) (*string, error)
}
