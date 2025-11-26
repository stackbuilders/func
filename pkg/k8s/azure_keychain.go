package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type AzureToken struct {
	Token     string `json:"token"`      // Azure access token
	ExpiresOn int64  `json:"expires_on"` // The expiration time of the token in Unix format
}

func GetAzureToken(configPath string, registry string) (string, error) {
	// Retrieves an Azure token for the specified registry.
	// If a valid cached token exists, it is returned. Otherwise, a new token is fetched and cached.
	cachedToken, err := ReadAzureToken(configPath, registry)
	if err != nil {
		return "", fmt.Errorf("failed to read Azure token: %w", err)
	}

	if cachedToken == "" {
		azCredential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return "", fmt.Errorf("failed to create default Azure credentials: %v", err)
		}
		// Define the default scope for Azure token requests
		defaultScope := "https://management.azure.com/.default"
		aztoken, err := azCredential.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{defaultScope}})
		if err != nil {
			return "", fmt.Errorf("failed to get Azure access token: %v", err)
		}
		SaveAzureToken(registry, aztoken)
		return aztoken.Token, nil
	}

	return cachedToken, nil
}

func TokenFile(configPath string, registry string) string {
	return filepath.Join(configPath, fmt.Sprintf("azure-token-%s.json", registry))
}

func ReadAzureToken(configPath string, registry string) (string, error) {
	// Returns an empty string if the token does not exist or is expired.
	tokenFile := TokenFile(configPath, registry)
	if _, err := os.Stat(tokenFile); err != nil {
		if os.IsNotExist(err) {
			return "", nil // No token file found, return empty token
		}
		return "", fmt.Errorf("failed to stat Azure token file: %w", err)
	}

	tokenBytes, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read Azure token file: %w", err)
	}

	var token AzureToken
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return "", fmt.Errorf("failed to unmarshal Azure token: %w", err)
	}

	if time.Now().After(time.Unix(0, token.ExpiresOn)) {
		// Token has expired, return empty token
		return "", nil
	}

	return token.Token, nil
}

func SaveAzureToken(configPath string, registry string, token azcore.AccessToken) error {
	// Saves an Azure token to the file system for caching.
	tokenBytes, err := json.Marshal(AzureToken{
		Token:     token.Token,
		ExpiresOn: token.ExpiresOn.Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal Azure token: %w", err)
	}

	tokenFile := TokenFile(configPath, registry)
	// Ensure the directory for the token file exists
	if err := os.MkdirAll(filepath.Dir(tokenFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory for Azure token file: %w", err)
	}

	if err := os.WriteFile(tokenFile, tokenBytes, 0600); err != nil {
		return fmt.Errorf("failed to write Azure token file: %w", err)
	}

	return nil
}
