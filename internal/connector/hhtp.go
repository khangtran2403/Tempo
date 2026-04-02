package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPConnector struct {
	client         *http.Client
	circuitbreaker *CircuitBreaker
}

func NewHTTPConnector() *HTTPConnector {
	return &HTTPConnector{
		client: &http.Client{
			Timeout: 40 * time.Second,
		},
		circuitbreaker: NewCircuitBreaker(6, 30*time.Second),
	}
}
func (h *HTTPConnector) Name() string {
	return "http"
}

func (h *HTTPConnector) Execute(ctx context.Context, config map[string]interface{}, prevRes map[string]interface{}) (interface{}, error) {
	method, ok := config["method"].(string)
	if !ok {
		return nil, fmt.Errorf("hhtp:Thiếu hoặc method không hợp lệ")
	}
	url, ok := config["url"].(string)
	if !ok {
		return nil, fmt.Errorf("Thiếu url")
	}
	headers := make(map[string]string)
	if h, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}
	var bodyReader io.Reader
	if body, ok := config["body"]; ok {
		// Support cả string và map
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case map[string]interface{}:
			jsonBody, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("http: chuyển đổi định dạng dữ liệu thất bại: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)

			if _, exists := headers["Content-Type"]; !exists {
				headers["Content-Type"] = "application/json"
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Gửi yêu cầu thất bại: %v", err)
	}

	//set header
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	startTime := time.Now()
	resp, err := h.client.Do(req)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("http: không thể gửi yêu cầu: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http: đọc response thất bại: %w", err)
	}

	var responsejson interface{}
	if err := json.Unmarshal(respBody, &responsejson); err != nil {
		responsejson = string(respBody)
	}
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     resp.Header,
		"body":        responsejson,
		"duration_ms": duration,
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}
	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("http: yêu cầu thất bại với mã trạng thái %d", resp.StatusCode)
	}

	return result, nil
}

func (h *HTTPConnector) ValidateConfig(config map[string]interface{}) error {
	//Validate methods
	method, ok := config["method"].(string)
	if !ok || method == "" {
		return fmt.Errorf("http: 'phương thức' là bắt buộc")
	}

	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	isValid := false
	for _, m := range validMethods {
		if strings.ToUpper(method) == m {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("http: Phương thức không hợp lệ '%s'", method)
	}

	// Validate URL
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("http: 'url' là bắt buộc")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("http: url phải bắt đầu bằng http:// hoặc https://")
	}

	return nil
}
