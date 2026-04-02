package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/crypto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleDriveConnector struct {
	integrationRepo storage.IntegrationRepository
	encryptionKey   string
	cfg             *config.Config
}

func NewGoogleDriveConnector(
	integrationRepo storage.IntegrationRepository,
	encryptionKey string,
	cfg *config.Config,
) *GoogleDriveConnector {
	return &GoogleDriveConnector{
		integrationRepo: integrationRepo,
		encryptionKey:   encryptionKey,
		cfg:             cfg,
	}
}

func (g *GoogleDriveConnector) Name() string {
	return "google_drive"
}

func (g *GoogleDriveConnector) ValidateConfig(cfg map[string]interface{}) error {
	if _, ok := cfg["filename"].(string); !ok {
		return fmt.Errorf("google_drive: 'tên file' là bắt buộc")
	}
	_, contentOk := cfg["content"].(string)
	_, filePathOk := cfg["file_path"].(string)
	if !contentOk && !filePathOk {
		return fmt.Errorf("google_drive: hoặc 'nội dung' hoặc 'đường dẫn' là bắt buộc")
	}
	return nil
}

func (g *GoogleDriveConnector) Execute(ctx context.Context, config map[string]interface{}, prevResults map[string]interface{}) (interface{}, error) {
	// 1. Get authenticated drive service
	integrationID, _ := config["integration_id"].(string)
	if integrationID == "" {
		return nil, fmt.Errorf("google_drive: 'id tích hợp' is required")
	}

	driveService, err := g.getDriveService(ctx, integrationID)
	if err != nil {
		return nil, err
	}

	// 2. Get config values - they are already rendered by the activity
	filename, _ := config["filename"].(string)
	parentFolderId, _ := config["parent_folder_id"].(string)
	contentType, _ := config["content_type"].(string)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 3. Prepare file metadata
	fileMetadata := &drive.File{
		Name:     filename,
		MimeType: contentType,
	}
	if parentFolderId != "" {
		fileMetadata.Parents = []string{parentFolderId}
	}

	// 4. Prepare file content
	var contentReader io.Reader
	if filePath, ok := config["file_path"].(string); ok && filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("google_drive: mở file thất bại %s: %w", filePath, err)
		}
		defer file.Close()
		contentReader = file
	} else if content, ok := config["content"].(string); ok {
		contentReader = strings.NewReader(content)
	}

	if contentReader == nil {
		return nil, fmt.Errorf("google_drive: nội dung file trống không thể tải lên")
	}

	// 5. Create/Upload file
	file, err := driveService.Files.Create(fileMetadata).Media(contentReader).Do()
	if err != nil {
		return nil, fmt.Errorf("google_drive: không thể tạo fil: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"id":          file.Id,
		"name":        file.Name,
		"mimeType":    file.MimeType,
		"webViewLink": file.WebViewLink,
	}, nil
}
func (g *GoogleDriveConnector) getDriveService(ctx context.Context, integrationID string) (*drive.Service, error) {
	integration, err := g.integrationRepo.GetByID(ctx, integrationID)
	if err != nil || integration == nil {
		return nil, fmt.Errorf("google_drive: ID tích hợp '%s' không tồn tại", integrationID)
	}
	if integration.Provider != "google" {
		return nil, fmt.Errorf("google_drive: Tích hợp '%s' không phải là tích hợp Google", integrationID)
	}

	decryptedToken, err := crypto.Decrypt(integration.AccessToken, g.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("google_drive: không thể giải mã token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(decryptedToken), &token); err != nil {
		return nil, fmt.Errorf("google_drive: không thể giải mã token: %w", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     g.cfg.Google.ClientID,
		ClientSecret: g.cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
	}

	tokenSource := oauthConfig.TokenSource(ctx, &token)
	srv, err := drive.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("google_drive: không thể tạo dịch vụ drive: %w", err)
	}
	return srv, nil
}
