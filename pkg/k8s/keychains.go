package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"knative.dev/func/pkg/creds"
	"knative.dev/func/pkg/oci"
)

func GetGoogleCredentialLoader() []creds.CredentialsCallback {
	return []creds.CredentialsCallback{
		func(registry string) (oci.Credentials, error) {
			if registry != "gcr.io" {
				return oci.Credentials{}, nil // skip if not GCR
			}

			res, err := name.NewRegistry(registry)
			if err != nil {
				return oci.Credentials{}, fmt.Errorf("parse registry: %w", err)
			}

			authenticator, err := google.Keychain.Resolve(res)
			if err != nil {
				return oci.Credentials{}, fmt.Errorf("resolve google keychain: %w", err)
			}

			authCfg, err := authenticator.Authorization()
			if err != nil {
				return oci.Credentials{}, fmt.Errorf("get authorization: %w", err)
			}

			return oci.Credentials{
				Username: authCfg.Username,
				Password: authCfg.Password,
			}, nil
		},
	}
}

func GetECRCredentialLoader() []creds.CredentialsCallback {
	return []creds.CredentialsCallback{} // TODO: Implement ECR credentials loader
}

func GetACRCredentialLoader() []creds.CredentialsCallback {
	return []creds.CredentialsCallback{
		func(registry string) (oci.Credentials, error) {
			if !strings.HasSuffix(registry, ".azurecr.io") {
				return oci.Credentials{}, nil
			}

			azCredential, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
				return oci.Credentials{}, fmt.Errorf("failed to create default Azure credentials: %v", err)
			}

			scope := "https://containerregistry.azure.net/.default"
			token, err := azCredential.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{scope}})
			if err != nil {
				return oci.Credentials{}, fmt.Errorf("failed to get Azure access token: %v", err)
			}

			return oci.Credentials{
				Username: "00000000-0000-0000-0000-000000000000",
				Password: token.Token,
			}, nil
		},
	}
}
