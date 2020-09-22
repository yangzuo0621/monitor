package akv

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

type akvClient struct {
	clientID     string
	tenantID     string
	clientSecret string
	vaultName    string
}

const (
	vaultURL = "https://%s.vault.azure.net"
	patKey   = "VSTSPAT"
)

// BuildAKVClient build a AKV client instance
func BuildAKVClient(clientID string, tenantID string, clientSecret string, vaultName string) AKVClient {
	return &akvClient{
		clientID:     clientID,
		tenantID:     tenantID,
		clientSecret: clientSecret,
		vaultName:    vaultName,
	}
}

// GetAzureKeyVaultAuthorizer constructs authorizer for AKV
func (c *akvClient) GetAzureKeyVaultAuthorizer() (autorest.Authorizer, error) {
	cloudEnv, err := Environment()
	if err != nil {
		return nil, err
	}
	alternateEndpoint, err := url.Parse(cloudEnv.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, err
	}
	alternateEndpoint.Path = path.Join(c.tenantID, "/oauth2/token")

	oauthConfig, err := adal.NewOAuthConfig(cloudEnv.ActiveDirectoryEndpoint, c.tenantID)
	if err != nil {
		return nil, fmt.Errorf("create OAuth config failed")
	}
	oauthConfig.AuthorizeEndpoint = *alternateEndpoint

	vaultEndpoint := strings.TrimSuffix(cloudEnv.KeyVaultEndpoint, "/")
	token, err := adal.NewServicePrincipalToken(*oauthConfig, c.clientID, c.clientSecret, vaultEndpoint)
	if err != nil {
		return nil, fmt.Errorf("create service principal token failed: %w", err)
	}

	return autorest.NewBearerAuthorizer(token), nil
}

func (c *akvClient) GetSecretFromAzureKeyVault(ctx context.Context, secretName string) (*string, error) {
	authorizer, err := c.GetAzureKeyVaultAuthorizer()
	if err != nil {
		return nil, err
	}

	kvClient := keyvault.New()
	kvClient.Authorizer = authorizer

	secret, err := kvClient.GetSecret(ctx, fmt.Sprintf(vaultURL, c.vaultName), secretName, "")
	if err != nil {
		return nil, err
	}

	return secret.Value, nil
}

func (c *akvClient) GetPAT(ctx context.Context) (string, error) {
	pat, err := c.GetSecretFromAzureKeyVault(ctx, patKey)
	return *pat, err
}

var _ AKVClient = (*akvClient)(nil)
