package k8s

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestACRCredentialLoader_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	loader := GetACRCredentialLoader()[0]
	registry := "example.azurecr.io"

	// Set dummy environment variables to ensure the loader attempts to authenticate
	// it will fail due to the context cancellation, but we want to ensure it fails for the right reason
	os.Setenv("AZURE_TENANT_ID", "dummy-tenant-id")
	os.Setenv("AZURE_CLIENT_ID", "dummy-client-id")
	os.Setenv("AZURE_CLIENT_SECRET", "dummy-client-secret")
	_, err := loader(ctx, registry)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Successfully caught context cancellation error: %v", err)
}
