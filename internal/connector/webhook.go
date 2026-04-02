// internal/connector/webhook.go
package connector

import (
	"context"
	"fmt"
)

// WebhookConnector xử lý webhook triggers
type WebhookConnector struct{}

func NewWebhookConnector() *WebhookConnector {
	return &WebhookConnector{}
}

func (w *WebhookConnector) Name() string {
	return "webhook"
}

// Execute - Webhook là trigger, không cần execute logic phức tạp
// Data đã được passed vào workflow từ HTTP request
func (w *WebhookConnector) Execute(
	ctx context.Context,
	config map[string]interface{},
	prevResults map[string]interface{},
) (interface{}, error) {

	triggerData, ok := config["trigger_data"]
	if !ok {
		return nil, fmt.Errorf("webhook: thiếu dữ liệu để kích hoạt(kiểm tra 'trigger_data')")
	}

	// Return trigger data để actions sau có thể dùng
	return map[string]interface{}{
		"trigger_type": "webhook",
		"data":         triggerData,
		"timestamp":    config["timestamp"],
	}, nil
}

func (w *WebhookConnector) ValidateConfig(config map[string]interface{}) error {
	// Webhook không cần config đặc biệt
	return nil
}
