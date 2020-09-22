package akv

import "github.com/Azure/go-autorest/autorest/azure"

var (
	cloudName   string = "AzurePublicCloud"
	environment *azure.Environment
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
