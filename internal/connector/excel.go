package connector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/pkg/common"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type ExcelConnector struct {
	outputPath string
	logger     *logrus.Logger
}

func NewExcelConnector(cfg *config.Config) *ExcelConnector {
	path := cfg.ExcelOutputPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
	return &ExcelConnector{
		outputPath: path,
		logger:     common.GetLogger(),
	}
}

func (e *ExcelConnector) Name() string {
	return "excel"
}

func (e *ExcelConnector) ValidateConfig(cfg map[string]interface{}) error {
	if _, ok := cfg["filename"].(string); !ok {
		return fmt.Errorf("excel: 'tên file' là bắt buộc và phải là chuỗi")
	}
	if _, ok := cfg["data"]; !ok {
		return fmt.Errorf("excel: 'dữ liệu' là bắt buộc")
	}
	return nil
}

func (e *ExcelConnector) Execute(ctx context.Context, config map[string]interface{}, prevResults map[string]interface{}) (interface{}, error) {
	// 1. Get config values and add EXTREME debugging.
	e.logger.Infof("[EXCEL DEBUG] Received config: %v", config)

	filename, _ := config["filename"].(string)
	data, exists := config["data"]
	if !exists {
		return nil, fmt.Errorf("excel: thiếu trường 'dữ liệu' trong cấu hình")
	}

	e.logger.Infof("[EXCEL DEBUG] 'data' field value: %#v", data)
	e.logger.Infof("[EXCEL DEBUG] 'data' field type: %T", data)

	// 2. Ensure data is a slice of maps
	dataSlice, ok := data.([]interface{})
	if !ok {
		e.logger.Errorf("[EXCEL DEBUG] Type assertion data.([]interface{}) failed. The data is not a slice as expected.")
		return nil, fmt.Errorf("excel: Trường 'dữ liệu' phải là một mảng các đối tượng")
	}
	if len(dataSlice) == 0 {
		return map[string]interface{}{"status": "skipped", "reason": "no data to write"}, nil
	}

	// 3. Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	streamWriter, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return nil, fmt.Errorf("excel: không thể tạo file excel: %w", err)
	}

	// 4. Get headers
	var headers []string
	if h, ok := config["headers"].([]interface{}); ok && len(h) > 0 {
		for _, header := range h {
			headers = append(headers, fmt.Sprintf("%v", header))
		}
	} else {
		if len(dataSlice) > 0 {
			if firstItem, ok := dataSlice[0].(map[string]interface{}); ok {
				for key := range firstItem {
					headers = append(headers, key)
				}
			}
		}
	}

	// 5. Write Header Row
	if len(headers) > 0 {
		headerRow := make([]interface{}, len(headers))
		for i, h := range headers {
			headerRow[i] = h
		}
		if err := streamWriter.SetRow("A1", headerRow); err != nil {
			return nil, fmt.Errorf("excel: không thể ghi hàng tiêu đề: %w", err)
		}
	}

	// 6. Write Data Rows
	for i, item := range dataSlice {
		row := make([]interface{}, len(headers))
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		for j, header := range headers {
			if val, exists := itemMap[header]; exists {
				row[j] = val
			} else {
				row[j] = ""
			}
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := streamWriter.SetRow(cell, row); err != nil {
			return nil, fmt.Errorf("excel: không thể ghi hàng dữ liệu %d: %w", i+1, err)
		}
	}

	// 7. Save file
	if err := streamWriter.Flush(); err != nil {
		return nil, fmt.Errorf("excel: không thể lưu file: %w", err)
	}

	fullPath := filepath.Join(e.outputPath, sanitizeFilename(filename))
	if err := f.SaveAs(fullPath); err != nil {
		return nil, fmt.Errorf("excel: không thể lưu file: %w", err)
	}

	return map[string]interface{}{
			"success":   true,
			"file_path": fullPath,
			"rows":      len(dataSlice),
		},
		nil
}

func sanitizeFilename(filename string) string {
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	return filename
}
