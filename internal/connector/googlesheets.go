package connector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/crypto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsConnector struct {
	integrationRepo storage.IntegrationRepository
	encryptionKey   string
	cfg             *config.Config
}

func NewGoogleSheetsConnector(
	integrationRepo storage.IntegrationRepository,
	encryptionKey string,
	cfg *config.Config,
) *GoogleSheetsConnector {
	return &GoogleSheetsConnector{
		integrationRepo: integrationRepo,
		encryptionKey:   encryptionKey,
		cfg:             cfg,
	}
}

func (g *GoogleSheetsConnector) Name() string {
	return "google_sheets"
}

func (g *GoogleSheetsConnector) ValidateConfig(cfg map[string]interface{}) error {
	if _, ok := cfg["spreadsheet_id"].(string); !ok {
		return fmt.Errorf("google_sheets: 'spreadsheet_id' là bắt buộc")
	}
	if _, ok := cfg["row_data"]; !ok { // Check for existence, not type string
		return fmt.Errorf("google_sheets: 'row_data' là bắt buộc")
	}
	return nil
}

func (g *GoogleSheetsConnector) Execute(ctx context.Context, config map[string]interface{}, prevResults map[string]interface{}) (interface{}, error) {
	// 1. Get authenticated sheets service
	integrationID, _ := config["integration_id"].(string)
	if integrationID == "" {
		return nil, fmt.Errorf("google_sheets: 'id tích hợp' là bắt buộc")
	}

	sheetsService, err := g.getSheetsService(ctx, integrationID)
	if err != nil {
		return nil, err
	}

	// 2. Access config values directly (they are already rendered)
	spreadsheetId, _ := config["spreadsheet_id"].(string)
	sheetName, _ := config["sheet_name"].(string)
	rowData, exists := config["row_data"]
	if !exists {
		return nil, fmt.Errorf("google_sheets: thiếu 'row_data'")
	}

	// 3. Prepare data for insertion
	var values []interface{}
	if dataSlice, ok := rowData.([]interface{}); ok {
		values = dataSlice
	} else {
		return nil, fmt.Errorf("google_sheets: row_data không hợp lệ, nó phải là một mảng")
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	// 4. Append to sheet
	rangeStr := sheetName
	if rangeStr == "" {
		return nil, fmt.Errorf("google_sheets: 'sheet_name' là bắt buộc")
	}

	appendCall := sheetsService.Spreadsheets.Values.Append(spreadsheetId, rangeStr, valueRange)
	appendCall.ValueInputOption("USER_ENTERED")
	res, err := appendCall.Do()
	if err != nil {
		return nil, fmt.Errorf("google_sheets: không thể thêm dữ liệu: %w", err)
	}

	return map[string]interface{}{
			"success":      true,
			"spreadsheet":  res.SpreadsheetId,
			"updatedRange": res.Updates.UpdatedRange,
			"updatedRows":  res.Updates.UpdatedRows,
		},
		nil
}

func (g *GoogleSheetsConnector) getSheetsService(ctx context.Context, integrationID string) (*sheets.Service, error) {
	integration, err := g.integrationRepo.GetByID(ctx, integrationID)
	if err != nil || integration == nil {
		return nil, fmt.Errorf("google_sheets: ID tích hợp '%s' không tồn tại", integrationID)
	}

	if integration.Provider != "google" {
		return nil, fmt.Errorf("google_sheets: tích hợp '%s' không phải là tích hợp Google", integrationID)
	}

	decryptedToken, err := crypto.Decrypt(integration.AccessToken, g.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("google_sheets: không thể giải mã token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(decryptedToken), &token); err != nil {
		return nil, fmt.Errorf("google_sheets: không thể giải mã token: %w", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     g.cfg.Google.ClientID,
		ClientSecret: g.cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
	}

	tokenSource := oauthConfig.TokenSource(ctx, &token)
	srv, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("google_sheets: không thể tạo dịch vụ để kết nối google sheets: %w", err)
	}

	return srv, nil
}
