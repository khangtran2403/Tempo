package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	gcs "cloud.google.com/go/storage"
	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/crypto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type GCSConnector struct {
	integrationRepo storage.IntegrationRepository
	encryptionKey   string
	cfg             *config.Config
}

func NewGCSConnector(
	integrationRepo storage.IntegrationRepository,
	encryptionKey string,
	cfg *config.Config,
) *GCSConnector {
	return &GCSConnector{
		integrationRepo: integrationRepo,
		encryptionKey:   encryptionKey,
		cfg:             cfg,
	}
}

func (g *GCSConnector) Name() string {
	return "gcs"
}

func (g *GCSConnector) ValidateConfig(cfg map[string]interface{}) error {
	if _, ok := cfg["bucket_name"].(string); !ok {
		return fmt.Errorf("gcs: 'bucket_name' is required")
	}
	if _, ok := cfg["object_name"].(string); !ok {
		return fmt.Errorf("gcs: 'object_name' is required")
	}
	_, contentOk := cfg["content"].(string)
	_, filePathOk := cfg["file_path"].(string)
	if !contentOk && !filePathOk {
		return fmt.Errorf("gcs: either 'content' or 'file_path' is required")
	}
	return nil
}

func (g *GCSConnector) Execute(ctx context.Context, config map[string]interface{}, prevResults map[string]interface{}) (interface{}, error) {
	// 1. Get authenticated storage client
	integrationID, _ := config["integration_id"].(string)
	if integrationID == "" {
		return nil, fmt.Errorf("gcs: 'integration_id' is required")
	}
	storageClient, err := g.getStorageClient(ctx, integrationID)
	if err != nil {
		return nil, err
	}
	defer storageClient.Close()

	// 2. Get config values - they are already rendered
	bucketName, _ := config["bucket_name"].(string)
	objectName, _ := config["object_name"].(string)

	// 3. Get content to upload
	obj := storageClient.Bucket(bucketName).Object(objectName)
	wc := obj.NewWriter(ctx)

	if filePath, ok := config["file_path"].(string); ok && filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("gcs: failed to open local file at %s: %w", filePath, err)
		}
		defer file.Close()
		if _, err = io.Copy(wc, file); err != nil {
			return nil, fmt.Errorf("gcs: failed to copy file content to GCS: %w", err)
		}
	} else if content, ok := config["content"].(string); ok {
		if _, err = wc.Write([]byte(content)); err != nil {
			return nil, fmt.Errorf("gcs: failed to write content to GCS: %w", err)
		}
	}

	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("gcs: failed to close GCS writer: %w", err)
	}

	if err := obj.ACL().Set(ctx, gcs.AllUsers, gcs.RoleReader); err != nil {
		return nil, fmt.Errorf("gcs: failed to set public ACL: %w", err)
	}

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: failed to get object attributes: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"bucket":     attrs.Bucket,
		"object":     attrs.Name,
		"size":       attrs.Size,
		"media_link": attrs.MediaLink,
	}, nil
}

func (g *GCSConnector) getStorageClient(ctx context.Context, integrationID string) (*gcs.Client, error) {
	integration, err := g.integrationRepo.GetByID(ctx, integrationID)
	if err != nil || integration == nil {
		return nil, fmt.Errorf("gcs: integration with ID '%s' not found", integrationID)
	}
	if integration.Provider != "google" {
		return nil, fmt.Errorf("gcs: integration '%s' is not a Google integration", integrationID)
	}

	decryptedToken, err := crypto.Decrypt(integration.AccessToken, g.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("gcs: failed to decrypt token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(decryptedToken), &token); err != nil {
		return nil, fmt.Errorf("gcs: failed to unmarshal token: %w", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     g.cfg.Google.ClientID,
		ClientSecret: g.cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gcs.ScopeReadWrite},
	}

	tokenSource := oauthConfig.TokenSource(ctx, &token)
	client, err := gcs.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("gcs: failed to create storage client: %w", err)
	}

	return client, nil
}
