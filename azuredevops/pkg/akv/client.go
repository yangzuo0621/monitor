package akv

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

type AKVClient struct {
	ClientID     string
	TenantID     string
	ClientSecret string
}

// GetAzureKeyVaultAuthorizer constructs authorizer for AKV
func (c *AKVClient) GetAzureKeyVaultAuthorizer() (autorest.Authorizer, error) {
	cloudEnv, err := Environment()
	if err != nil {
		return nil, err
	}
	alternateEndpoint, err := url.Parse(cloudEnv.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, err
	}
	alternateEndpoint.Path = path.Join(c.TenantID, "/oauth2/token")

	oauthConfig, err := adal.NewOAuthConfig(cloudEnv.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		return nil, fmt.Errorf("create OAuth config failed")
	}
	oauthConfig.AuthorizeEndpoint = *alternateEndpoint

	vaultEndpoint := strings.TrimSuffix(cloudEnv.KeyVaultEndpoint, "/")
	token, err := adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, vaultEndpoint)
	if err != nil {
		return nil, fmt.Errorf("create service principal token failed: %w", err)
	}

	return autorest.NewBearerAuthorizer(token), nil
}
